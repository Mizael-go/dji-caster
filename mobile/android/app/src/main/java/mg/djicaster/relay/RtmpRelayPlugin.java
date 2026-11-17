package mg.djicaster.relay;

import android.Manifest;
import android.util.DisplayMetrics;
import android.view.Gravity;
import android.view.ViewGroup;
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
 */
@CapacitorPlugin(
    name = "RtmpRelay",
    permissions = {
      @Permission(alias = "camera", strings = { Manifest.permission.CAMERA }),
      @Permission(alias = "microphone", strings = { Manifest.permission.RECORD_AUDIO })
    })
public class RtmpRelayPlugin extends Plugin {

  private static final float PREVIEW_RATIO = 0.42f;

  private OpenGlView preview;
  private RtmpCamera2 camera;
  private String currentUrl;

  private final ConnectChecker connectChecker = new ConnectChecker() {
    @Override
    public void onConnectionStarted(String url) {
      emit("connecting", url);
    }

    @Override
    public void onConnectionSuccess() {
      emit("live", currentUrl);
    }

    @Override
    public void onConnectionFailed(@NonNull String reason) {
      emit("error", reason);
    }

    @Override
    public void onDisconnect() {
      emit("idle", null);
    }

    @Override
    public void onAuthError() {
      emit("error", "Authentification refusée par le serveur RTMP.");
    }

    @Override
    public void onAuthSuccess() {}

    @Override
    public void onNewBitrate(long bitrate) {}
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
    int width = call.getInt("width", 1280);
    int height = call.getInt("height", 720);
    int fps = call.getInt("fps", 30);
    int bitrate = call.getInt("bitrate", 2500000);
    String facing = call.getString("facing", "back");

    getActivity()
        .runOnUiThread(
            () -> {
              try {
                release();
                currentUrl = url;
                preview = new OpenGlView(getActivity());
                preview.setKeepScreenOn(true);
                FrameLayout content = getActivity().findViewById(android.R.id.content);
                DisplayMetrics metrics = getActivity().getResources().getDisplayMetrics();
                int previewHeight = (int) (metrics.heightPixels * PREVIEW_RATIO);
                content.addView(
                    preview,
                    new FrameLayout.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT, previewHeight, Gravity.TOP));

                camera = new RtmpCamera2(preview, connectChecker);
                if (!camera.prepareVideo(width, height, bitrate, fps, 90,
                        metrics.densityDpi)) {
                  throw new IllegalStateException("Encodage vidéo indisponible.");
                }
                camera.prepareAudio(128 * 1024, 44100, true);
                CameraHelper.Facing direction =
                    "front".equals(facing) ? CameraHelper.Facing.FRONT : CameraHelper.Facing.BACK;
                camera.startPreview(direction);
                camera.startStream(url);

                JSObject ret = new JSObject();
                ret.put("url", url);
                call.resolve(ret);
              } catch (Exception error) {
                release();
                call.reject("Diffusion impossible : " + error.getMessage());
              }
            });
  }

  @PluginMethod
  public void stop(PluginCall call) {
    getActivity()
        .runOnUiThread(
            () -> {
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

  @Override
  protected void handleOnDestroy() {
    release();
  }

  /** Arrête le flux et libère la vue d'aperçu. Doit tourner sur le thread UI. */
  private void release() {
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
