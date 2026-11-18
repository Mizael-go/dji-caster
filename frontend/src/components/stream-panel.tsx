import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Check, Copy, Loader2, Play, Power, Square, Waves, Zap } from "lucide-react"
import { Badge } from "~/components/ui/badge"
import { Button } from "~/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select"
import { api, copyText, errorMessage } from "~/lib/api"
import { cn, formatBytes } from "~/lib/utils"
import type { PathView, Snapshot } from "~/lib/types"

function PathRow({ label, path }: { label: string; path: PathView }) {
  return (
    <div className="flex items-center justify-between gap-2 rounded-xl bg-muted/60 px-2.5 py-2">
      <div className="flex min-w-0 items-center gap-2">
        <span
          className={cn(
            "size-1.5 shrink-0 rounded-full",
            path.ready ? "bg-[color:var(--color-signal)]" : "bg-muted-foreground/30",
          )}
        />
        <span className="truncate text-[11px] font-medium">{label}</span>
      </div>
      <div className="flex shrink-0 items-center gap-2 font-mono text-[9.5px] text-muted-foreground">
        <span>{path.ready ? (path.codecs ?? []).join(" · ") || "actif" : "—"}</span>
        <span className="tabular-nums">{formatBytes(path.bytes)}</span>
        <span className="tabular-nums">{path.readers} lect.</span>
      </div>
    </div>
  )
}

export function StreamPanel({ snapshot }: { snapshot: Snapshot }) {
  const queryClient = useQueryClient()
  const stream = snapshot.stream
  const [hostMode, setHostMode] = useState<"lan" | "local">("lan")
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState("")

  const url = hostMode === "local" ? stream.localURL : stream.obsURL
  const refresh = () => queryClient.invalidateQueries({ queryKey: ["snapshot"] })

  const serverMutation = useMutation({
    mutationFn: () => (snapshot.server.running ? api.stopServer() : api.startServer()),
    onSuccess: () => {
      setError("")
      void refresh()
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const transcodeMutation = useMutation({
    mutationFn: (enabled: boolean) => api.setTranscode(enabled),
    onSuccess: () => {
      setError("")
      void refresh()
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const onCopy = async () => {
    const ok = await copyText(url)
    setCopied(ok)
    window.setTimeout(() => setCopied(false), 1600)
  }

  return (
    <div className="space-y-3 pb-2">
      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="flex items-center gap-1.5 text-[12px] font-bold tracking-tight">
              <Waves className="size-3.5 text-[color:var(--color-signal)]" />
              Sortie RTMP
            </h2>
            <p className="mt-0.5 text-[11px] text-muted-foreground">À coller dans OBS</p>
          </div>
          <Badge
            variant="outline"
            className={cn(
              "h-5 shrink-0 gap-1 border px-1.5 text-[8.5px] font-bold uppercase tracking-[0.08em]",
              stream.incoming.ready
                ? "border-[color:var(--color-signal)]/35 bg-[color:var(--color-volt)]/25 text-[#3f4a1a]"
                : "border-border text-muted-foreground",
            )}
          >
            <span
              className={cn(
                "size-1.5 rounded-full",
                stream.incoming.ready
                  ? "bg-[color:var(--color-live)] live-dot"
                  : "bg-muted-foreground/40",
              )}
            />
            {stream.incoming.ready ? "flux actif" : "hors ligne"}
          </Badge>
        </div>

        <div className="mt-2.5 rounded-xl bg-[color:var(--color-ink)] p-2.5">
          <p className="select-all break-all font-mono text-[10.5px] leading-relaxed text-[color:var(--color-volt)]">
            {url}
          </p>
        </div>

        <div className="mt-2 flex items-center gap-1.5">
          <Select value={hostMode} onValueChange={(v) => setHostMode(v as "lan" | "local")}>
            <SelectTrigger size="sm" className="h-9 min-w-0 flex-1 bg-white">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="max-w-[260px]">
              <SelectItem value="lan">Réseau · {stream.host}</SelectItem>
              <SelectItem value="local">Local · 127.0.0.1</SelectItem>
            </SelectContent>
          </Select>
          <Button size="sm" onClick={onCopy} className="h-9 shrink-0">
            {copied ? <Check /> : <Copy />}
            {copied ? "Copié" : "Copier"}
          </Button>
        </div>

        <div className="mt-2.5 space-y-1.5">
          <PathRow label="Caméra → MediaMTX" path={stream.incoming} />
          {stream.transcode && <PathRow label="ffmpeg → H.264" path={stream.outgoing} />}
        </div>

        <div className="mt-2.5 rounded-xl border border-border bg-muted/40 px-2.5 py-2">
          <p className="text-[10px] font-semibold text-muted-foreground">
            Source à saisir dans DJI Relay
          </p>
          <p className="mt-1 select-all break-all font-mono text-[10px] text-foreground">
            {stream.directURL}
          </p>
          <p className="mt-1 text-[10px] leading-snug text-muted-foreground">
            Le téléphone publie sur cette URL. Si la conversion est activée, l’URL OBS ci-dessus
            utilise séparément <span className="font-mono text-foreground">live/dji-h264</span>.
          </p>
          <p className="mt-1.5 border-t border-border pt-1.5 text-[10px] leading-snug text-muted-foreground">
            <span className="font-semibold text-foreground">Latence</span> — dans OBS, propriétés
            de la source média : tampon réseau à 0 Mo et « Reconnecter » activé.
          </p>
        </div>

        {stream.transcode && !stream.incoming.ready && stream.outgoing.ready && (
          <p className="mt-2 rounded-lg bg-[color:var(--color-warning)]/10 px-2 py-1.5 text-[10.5px] leading-snug text-[color:var(--color-warning)]">
            Le flux est actif sur la sortie H.264 mais aucune source n’arrive sur{" "}
            <span className="font-mono">live/dji</span>. Dans DJI Relay, remplacez le chemin par{" "}
            <span className="font-mono">live/dji</span> : <span className="font-mono">live/dji-h264</span>{" "}
            est réservé à la sortie vers OBS.
          </p>
        )}
      </section>

      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-start gap-3">
          <div className="min-w-0 flex-1">
            <h2 className="flex items-center gap-1.5 text-[12px] font-bold tracking-tight">
              <Zap className="size-3.5 text-[color:var(--color-signal)]" />
              Convertir en H.264
            </h2>
            <p className="mt-0.5 text-[11px] leading-snug text-muted-foreground">
              Republie la scène sur <span className="font-mono text-foreground">live/dji-h264</span>{" "}
              avec des timestamps assainis : c'est l'URL à donner à OBS (évite les
              images figées).
            </p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={stream.transcode}
            aria-label="Convertir en H.264"
            disabled={transcodeMutation.isPending || !snapshot.environment.ffmpegReady}
            onClick={() => transcodeMutation.mutate(!stream.transcode)}
            className={cn(
              "relative mt-0.5 h-6 w-11 shrink-0 rounded-full border transition-colors",
              stream.transcode
                ? "border-[color:var(--color-signal)]/40 bg-[color:var(--color-signal)]"
                : "border-input bg-muted",
              (transcodeMutation.isPending || !snapshot.environment.ffmpegReady) &&
                "cursor-not-allowed opacity-50",
            )}
          >
            <span
              className={cn(
                "absolute top-[3px] size-[16px] rounded-full bg-white shadow-sm transition-all",
                stream.transcode ? "left-[24px]" : "left-[3px]",
              )}
            />
          </button>
        </div>

        {!stream.transcode && (stream.incoming.codecs ?? []).includes("H265") && (
          <p className="mt-2 rounded-lg bg-[color:var(--color-volt)]/25 px-2 py-1.5 text-[10.5px] leading-snug text-[#3f4a1a]">
            La caméra émet en HEVC (H.265). Si OBS n'affiche rien, activez la conversion.
          </p>
        )}
        {!stream.transcode &&
          stream.incoming.ready &&
          (stream.incoming.codecs ?? []).includes("H264") && (
            <div className="mt-2 rounded-lg bg-[color:var(--color-warning)]/10 px-2 py-2 text-[10.5px] leading-snug text-[color:var(--color-warning)]">
              <p>
                Le relais Android peut envoyer des timestamps incorrects : en lecture directe,
                OBS peut afficher une image en retard ou se figer. Activez la republication
                horodatée, puis utilisez l’URL OBS <span className="font-mono">live/dji-h264</span>.
              </p>
              <Button
                size="sm"
                variant="outline"
                className="mt-2 h-7 bg-card text-foreground"
                disabled={transcodeMutation.isPending || !snapshot.environment.ffmpegReady}
                onClick={() => transcodeMutation.mutate(true)}
              >
                Activer la sortie H.264 pour OBS
              </Button>
            </div>
          )}
        {!snapshot.environment.ffmpegReady && (
          <p className="mt-2 rounded-lg bg-[color:var(--color-warning)]/10 px-2 py-1.5 text-[10.5px] leading-snug text-[color:var(--color-warning)]">
            ffmpeg absent : installez-le (apt install ffmpeg) pour activer la conversion.
          </p>
        )}
        {stream.transcode && stream.transcodeErr && (
          <p className="mt-2 rounded-lg bg-destructive/10 px-2 py-1.5 text-[10.5px] leading-snug text-destructive">
            {stream.transcodeErr}
          </p>
        )}
        {stream.transcode && !stream.outgoing.ready && snapshot.stream.incoming.ready && (
          <p className="mt-2 text-[10.5px] text-muted-foreground">
            ffmpeg démarre le transcodage…
          </p>
        )}
      </section>

      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-center justify-between gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <Power
              className={cn(
                "size-3.5 shrink-0",
                snapshot.server.running
                  ? "text-[color:var(--color-signal)]"
                  : "text-muted-foreground",
              )}
            />
            <span className="truncate text-[12px] font-bold tracking-tight">Serveur MediaMTX</span>
            <Badge variant="outline" className="h-4 shrink-0 px-1 font-mono text-[8.5px]">
              :{snapshot.server.rtmpPort}
            </Badge>
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={() => serverMutation.mutate()}
            disabled={serverMutation.isPending}
            className="shrink-0 bg-white"
          >
            {serverMutation.isPending ? (
              <Loader2 className="animate-spin" />
            ) : snapshot.server.running ? (
              <Square />
            ) : (
              <Play />
            )}
            {snapshot.server.running ? "Arrêter" : "Démarrer"}
          </Button>
        </div>

        <p className="mt-2 text-[11px] leading-snug text-muted-foreground">
          {snapshot.server.running
            ? "Le serveur écoute et attend la caméra. OBS peut ouvrir l'URL à tout moment."
            : "Démarrez le serveur avant de connecter la caméra : c'est sa cible RTMP."}
        </p>

        {(error || snapshot.server.err) && (
          <p className="mt-2 rounded-lg bg-destructive/10 px-2 py-1.5 text-[10.5px] leading-snug text-destructive">
            {error || snapshot.server.err}
          </p>
        )}
      </section>
    </div>
  )
}
