package mg.djicaster.relay;

import android.Manifest;
import android.content.res.Configuration;
import android.graphics.Color;
import android.graphics.drawable.GradientDrawable;
import android.os.Handler;
import android.os.Looper;
import android.util.TypedValue;
import android.view.Gravity;
import android.view.Surface;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.FrameLayout;

import androidx.annotation.NonNull;

import com.getcapacitor.JSObject;
import com.getcapacitor.PermissionState;
import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;
import com.getcapacitor.annotation.Permission;
import com.getcapacitor.annotation.PermissionCallback;
import com.pedro.common.ConnectChecker;
import com.pedro.encoder.input.video.CameraHelper;
import com.pedro.library.rtmp.RtmpCamera2;
import com.pedro.library.view.OpenGlView;

/**
 * Publie la caméra et le micro du téléphone en RTMP sur le MediaMTX de DJI
 * Caster. Du point de vue du poste de travail, le flux est identique à celui
 * d'une caméra DJI : H.264 + AAC sur le chemin demandé (live/dji).
 *
 * L'aperçu utilise un OpenGlView natif ajouté par-dessus la WebView, dans la
 * bande haute (42 %) que la page web laisse libre.
 *
 * Robustesse : une chute de connexion (broken pipe, Wi-Fi instable, serveur
 * momentanément absent) déclenche une reconnexion automatique avec backoff.
 * Une rotation de l'écran redémarre la diffusion pour que l'encodage suive
 * la nouvelle orientation (le manifeste déclare configChanges, l'activité
 * n'est donc pas recréée).
 */
@CapacitorPlugin(
    name = "RtmpRelay",
    permissions = {
      @Permission(alias = "camera", strings = { Manifest.permission.CAMERA }),
      @Permission(alias = "microphone", strings = { Manifest.permission.RECORD_AUDIO })
    })
public class RtmpRelayPlugin extends Plugin {

  /** Délai laissé à Android pour appliquer la nouvelle orientation (ms). */
  private static final long ROTATION_SETTLE_MS = 400L;
  /** Délai maximal d'attente de la surface GL après l'ajout de la vue (ms). */
  private static final long SURFACE_TIMEOUT_MS = 5000L;
  /** Intervalle entre deux vérifications de la surface (ms). */
  private static final long SURFACE_POLL_MS = 50L;
  /** Backoff de reconnexion, puis abandon (l'utilisateur relance à la main). */
  private static final long[] RETRY_DELAYS_MS = { 2000L, 4000L, 8000L, 16000L, 30000L };

  private OpenGlView preview;
  private Button stopButton;
  private RtmpCamera2 camera;
  private String currentUrl;
  private int width;
  private int height;
  private int fps;
  private int bitrate;
  private String facing;
  /** true tant que l'utilisateur veut diffuser (entre start et stop/abandon). */
  private boolean started;
  /** Incrémenté à chaque démarrage/arrêt pour invalider les attentes en cours. */
  private long startSession;
  private int retries;
  private boolean retryScheduled;
  private int lastOrientation = Configuration.ORIENTATION_UNDEFINED;

  private final Handler mainHandler = new Handler(Looper.getMainLooper());

  private final ConnectChecker connectChecker = new ConnectChecker() {
    @Override
    public void onConnectionStarted(String url) {
      emit("connecting", url);
    }

    @Override
    public void onConnectionSuccess() {
      retries = 0;
      emit("live", currentUrl);
    }

    @Override
    public void onConnectionFailed(@NonNull String reason) {
      scheduleReconnect(reason);
    }

    @Override
    public void onDisconnect() {
      // La socket est morte : RootEncoder appelle en général ensuite
      // onConnectionFailed, qui déclenche la reconnexion. On reflète
      // l'état pour que l'UI ne laisse pas croire que tout va bien.
      if (started) {
        emit("reconnecting", "Connexion perdue, reconnexion en cours…");
      } else {
        emit("idle", null);
      }
    }

    @Override
    public void onAuthError() {
      started = false;
      emit("error", "Authentification refusée par le serveur RTMP.");
    }

    @Override
    public void onAuthSuccess() {}

    @Override
    public void onNewBitrate(long bitrate) {}
  };

  private final Runnable retryRunnable = new Runnable() {
    @Override
    public void run() {
      getActivity().runOnUiThread(() -> retryStream());
    }
  };

  @PluginMethod
  public void start(PluginCall call) {
    if (getPermissionState("camera") != PermissionState.GRANTED
        || getPermissionState("microphone") != PermissionState.GRANTED) {
      requestPermissionForAliases(
          new String[] { "camera", "microphone" }, call, "afterPermissions");
      return;
    }
    startRelay(call);
  }

  @PermissionCallback
  private void afterPermissions(PluginCall call) {
    if (getPermissionState("camera") != PermissionState.GRANTED) {
      call.reject("Permission caméra refusée.");
      return;
    }
    startRelay(call);
  }

  private void startRelay(PluginCall call) {
    String url = call.getString("url");
    if (url == null || url.isEmpty()) {
      call.reject("URL RTMP manquante.");
      return;
    }
    currentUrl = url;
    width = call.getInt("width", 1280);
    height = call.getInt("height", 720);
    fps = call.getInt("fps", 30);
    bitrate = call.getInt("bitrate", 2500000);
    facing = call.getString("facing", "back");
    getActivity()
        .runOnUiThread(
            () -> {
              try {
                internalStart(call);
              } catch (Exception error) {
                release();
                started = false;
                call.reject("Diffusion impossible : " + error.getMessage());
              }
            });
  }

  /**
   * Séquence complète de démarrage (aperçu + flux) avec les paramètres
   * courant. Appelée au premier start, après une rotation, et au retour
   * des permissions. Doit tourner sur le thread UI.
   */
  private void internalStart(PluginCall call) {
    release();
    startSession++;
    long session = startSession;
    started = true;
    retries = 0;
    preview = new OpenGlView(getActivity());
    preview.setKeepScreenOn(true);
    FrameLayout content = getActivity().findViewById(android.R.id.content);
    // Plein écran : l'aperçu caméra remplace l'interface pendant la
    // diffusion. Un bouton natif « Arrêter » reste au-dessus, sinon la
    // WebView serait inaccessible sous la vue caméra.
    content.addView(
        preview,
        new FrameLayout.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.MATCH_PARENT,
            Gravity.FILL));

    stopButton = new Button(getActivity());
    stopButton.setText("Arrêter");
    stopButton.setAllCaps(false);
    stopButton.setTextColor(Color.WHITE);
    stopButton.setTextSize(TypedValue.COMPLEX_UNIT_SP, 14f);
    GradientDrawable pill = new GradientDrawable();
    pill.setColor(Color.parseColor("#CC0A0B0D"));
    pill.setCornerRadius(dp(24));
    stopButton.setBackground(pill);
    stopButton.setPadding(dp(20), dp(8), dp(20), dp(8));
    stopButton.setOnClickListener(v -> {
      started = false;
      release();
      emit("idle", null);
    });
    FrameLayout.LayoutParams stopParams = new FrameLayout.LayoutParams(
        ViewGroup.LayoutParams.WRAP_CONTENT,
        ViewGroup.LayoutParams.WRAP_CONTENT,
        Gravity.BOTTOM | Gravity.CENTER_HORIZONTAL);
    stopParams.bottomMargin = dp(32);
    content.addView(stopButton, stopParams);

    camera = new RtmpCamera2(preview, connectChecker);
    // Ordre des paramètres de Camera2Base (RootEncoder 2.8.2) :
    // prepareVideo(width, height, fps, bitrate, rotation) avec
    // rotation = orientation du capteur, jamais la densité d'écran.
    // RootEncoder permute largeur/hauteur quand la rotation vaut 90/270 :
    // en portrait on publie donc bien 720x1280 avec width=1280/height=720.
    if (!camera.prepareVideo(width, height, fps, bitrate,
            CameraHelper.getCameraOrientation(getActivity()))) {
      throw new IllegalStateException("Encodage vidéo indisponible.");
    }
    camera.prepareAudio(128 * 1024, 44100, true);
    CameraHelper.Facing direction =
        "front".equals(facing) ? CameraHelper.Facing.FRONT : CameraHelper.Facing.BACK;
    // La Surface de l'OpenGlView n'existe qu'après le premier
    // passage de layout : si on appelle startPreview tout de suite,
    // l'init GL échoue en silence (secureSubmit avale l'exception)
    // et getSurfaceTexture() renvoie null. On attend sa création.
    waitForSurface(direction, call, session,
        System.currentTimeMillis() + SURFACE_TIMEOUT_MS);
  }

  /**
   * Démarre l'aperçu puis le flux dès que la surface GL est prête, en
   * repoussant la vérification sur la file de la vue. Doit tourner sur le
   * thread UI.
   */
  private void waitForSurface(
      CameraHelper.Facing direction, PluginCall call, long session, long deadline) {
    if (session != startSession || camera == null || preview == null) {
      // stop() ou un nouveau démarrage est passé entre-temps : abandon.
      if (call != null) {
        call.reject("Diffusion annulée.");
      }
      return;
    }
    Surface surface = preview.getHolder().getSurface();
    if (surface == null || !surface.isValid()) {
      if (System.currentTimeMillis() >= deadline) {
        release();
        started = false;
        if (call != null) {
          call.reject("L'aperçu caméra n'a pas pu démarrer (surface indisponible).");
        } else {
          emit("error", "L'aperçu caméra n'a pas pu redémarrer après la rotation.");
        }
        return;
      }
      preview.postDelayed(
          () -> waitForSurface(direction, call, session, deadline), SURFACE_POLL_MS);
      return;
    }
    try {
      camera.startPreview(direction);
      camera.startStream(currentUrl);
      if (call != null) {
        JSObject ret = new JSObject();
        ret.put("url", currentUrl);
        call.resolve(ret);
      }
    } catch (Exception error) {
      release();
      started = false;
      if (call != null) {
        call.reject("Diffusion impossible : " + error.getMessage());
      } else {
        emit("error", "Diffusion impossible après la rotation : " + error.getMessage());
      }
    }
  }

  /**
   * Reconnexion après une chute de connexion : stoppe le flux côté client
   * (l'aperçu reste vivant) puis retente avec un backoff croissant.
   */
  private void scheduleReconnect(String reason) {
    if (!started) {
      // Échec de connexion alors que personne ne diffuse : simple erreur.
      emit("error", reason);
      return;
    }
    if (retryScheduled) {
      return; // une reconnexion est déjà programmée
    }
    if (retries >= RETRY_DELAYS_MS.length) {
      retryScheduled = false;
      started = false;
      release();
      emit("error",
          "Diffusion perdue après " + RETRY_DELAYS_MS.length
              + " tentatives : " + reason);
      return;
    }
    retryScheduled = true;
    long delay = RETRY_DELAYS_MS[retries];
    retries++;
    emit("reconnecting",
        "Reconnexion dans " + (delay / 1000) + "s (tentative " + retries + "/"
            + RETRY_DELAYS_MS.length + ") — " + reason);
    mainHandler.postDelayed(retryRunnable, delay);
  }

  /** Retente la connexion sans toucher à l'aperçu. Thread UI. */
  private void retryStream() {
    retryScheduled = false;
    long session = startSession;
    if (!started || session != startSession || camera == null) {
      return; // stop() ou redémarrage passés entre-temps
    }
    try {
      if (camera.isStreaming()) {
        camera.stopStream();
      }
      camera.startStream(currentUrl);
    } catch (Exception error) {
      scheduleReconnect(error.getMessage());
    }
  }

  @PluginMethod
  public void stop(PluginCall call) {
    getActivity()
        .runOnUiThread(
            () -> {
              started = false;
              release();
              emit("idle", null);
              call.resolve();
            });
  }

  @PluginMethod
  public void status(PluginCall call) {
    JSObject ret = new JSObject();
    ret.put("streaming", camera != null && camera.isStreaming());
    ret.put("url", currentUrl);
    call.resolve(ret);
  }

  /**
   * Rotation de l'écran : l'activité n'est pas recréée (configChanges dans
   * le manifeste), on redémarre donc la diffusion pour que l'encodage
   * produise les dimensions de la nouvelle orientation.
   */
  @Override
  protected void handleOnConfigurationChanged(Configuration newConfig) {
    super.handleOnConfigurationChanged(newConfig);
    if (newConfig.orientation == lastOrientation || !started) {
      return;
    }
    lastOrientation = newConfig.orientation;
    getActivity()
        .runOnUiThread(
            () -> {
              if (!started) {
                return;
              }
              // On laisse Android finir d'appliquer la rotation (la rotation
              // d'affichage lue par getCameraOrientation doit être à jour)
              // avant de relancer l'encodeur dans les bonnes dimensions.
              final long session = startSession;
              mainHandler.postDelayed(
                  () -> {
                    if (!started || session != startSession) {
                      return;
                    }
                    internalStart(null);
                  },
                  ROTATION_SETTLE_MS);
            });
  }

  @Override
  protected void handleOnDestroy() {
    started = false;
    release();
  }

  /** Arrête le flux et libère la vue d'aperçu. Doit tourner sur le thread UI. */
  private void release() {
    startSession++;
    mainHandler.removeCallbacks(retryRunnable);
    retryScheduled = false;
    try {
      if (camera != null) {
        if (camera.isStreaming()) {
          camera.stopStream();
        }
        camera.stopPreview();
        camera = null;
      }
    } catch (Exception ignored) {
      // un arrêt en erreur ne doit jamais bloquer la fermeture de l'aperçu
    }
    if (preview != null) {
      ViewGroup parent = (ViewGroup) preview.getParent();
      if (parent != null) {
        parent.removeView(preview);
      }
      preview = null;
    }
    if (stopButton != null) {
      ViewGroup parent = (ViewGroup) stopButton.getParent();
      if (parent != null) {
        parent.removeView(stopButton);
      }
      stopButton = null;
    }
  }

  /** Convertit des dp en pixels. */
  private int dp(int value) {
    return Math.round(TypedValue.applyDimension(
        TypedValue.COMPLEX_UNIT_DIP, value,
        getActivity().getResources().getDisplayMetrics()));
  }

  private void emit(String state, String message) {
    JSObject data = new JSObject();
    data.put("state", state);
    if (message != null) {
      data.put("message", message);
    }
    notifyListeners("state", data);
  }
}
