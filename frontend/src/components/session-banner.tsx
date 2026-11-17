import { AlertTriangle, Loader2, Radio, X } from "lucide-react"
import { Button } from "~/components/ui/button"
import { api } from "~/lib/api"
import { cn } from "~/lib/utils"
import type { SessionView } from "~/lib/types"

export function SessionBanner({
  session,
  onReset,
}: {
  session: SessionView
  onReset: () => void
}) {
  if (session.state === "idle") return null

  const live = session.state === "live"
  const error = session.state === "error"
  const starting = session.state === "starting"

  const title = live
    ? `Flux reçu · ${session.deviceName || "DJI"}`
    : error
      ? "Connexion DJI échouée"
      : starting
        ? `Négociation BLE · ${session.deviceName || session.deviceAddr}`
        : "Ordre envoyé · en attente du flux"

  const detail = live
    ? "La caméra pousse vers MediaMTX. L'URL RTMP est prête pour OBS."
    : error
      ? session.err
      : starting
        ? "djictl dialogue avec la caméra pour la basculer sur votre Wi-Fi."
        : "La caméra se connecte au Wi-Fi puis démarre son encodeur RTMP."

  return (
    <div
      className={cn(
        "mt-3 rounded-2xl border p-3",
        live && "border-[color:var(--color-signal)]/30 bg-[color:var(--color-volt)]/25",
        error && "border-destructive/30 bg-destructive/[0.06]",
        !live && !error && "border-border bg-card card-raised",
      )}
    >
      <div className="flex items-start gap-2">
        {error ? (
          <AlertTriangle className="mt-px size-3.5 shrink-0 text-destructive" />
        ) : starting || !live ? (
          <Loader2 className="mt-px size-3.5 shrink-0 animate-spin text-[color:var(--color-warning)]" />
        ) : (
          <Radio className="mt-px size-3.5 shrink-0 text-[color:var(--color-signal)]" />
        )}
        <div className="min-w-0 flex-1">
          <p className="text-[12.5px] font-semibold leading-snug">{title}</p>
          <p
            className={cn(
              "mt-0.5 text-[11px] leading-snug",
              live ? "text-[#3f4a1a]" : "text-muted-foreground",
            )}
          >
            {detail}
          </p>
          {session.hint && (
            <p className="mt-1 text-[11px] leading-snug text-destructive">{session.hint}</p>
          )}
        </div>
        <button
          type="button"
          onClick={starting ? () => void api.cancelConnect() : onReset}
          className="grid size-6 shrink-0 place-items-center rounded-full text-muted-foreground transition-colors hover:bg-black/5 hover:text-foreground"
          aria-label={starting ? "Annuler la connexion" : "Réinitialiser"}
        >
          <X className="size-3.5" />
        </button>
      </div>
      {starting && (
        <Button
          size="sm"
          variant="outline"
          className="mt-2.5 w-full"
          onClick={() => void api.cancelConnect()}
        >
          Annuler la connexion
        </Button>
      )}
    </div>
  )
}
