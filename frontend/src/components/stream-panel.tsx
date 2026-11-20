import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  Check,
  Copy,
  ExternalLink,
  Loader2,
  Play,
  Power,
  Radio,
  Square,
  Volume2,
  Waves,
} from "lucide-react"
import { Badge } from "~/components/ui/badge"
import { Button } from "~/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select"
import { api, copyText, errorMessage, openURL } from "~/lib/api"
import { cn, formatBytes } from "~/lib/utils"
import type { PathView, Snapshot } from "~/lib/types"

// MediaMTX nomme les pistes d'après leur codec : toute piste qui n'est pas
// vidéo est de l'audio (Opus, AAC…).
const VIDEO_CODECS = ["H264", "H265", "AV1", "VP8", "VP9", "MPEG-4 Video", "MPEG-1/2 Video", "M-JPEG"]
const hasAudio = (codecs: string[]) =>
  codecs.some((codec) => !VIDEO_CODECS.some((video) => codec.startsWith(video)))

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
  const [copiedKey, setCopiedKey] = useState("")
  const [error, setError] = useState("")

  const webrtcUrl = hostMode === "local" ? stream.webrtcLocalURL : stream.webrtcURL
  const incomingAudio = hasAudio(stream.incoming.codecs ?? [])
  const refresh = () => queryClient.invalidateQueries({ queryKey: ["snapshot"] })

  const serverMutation = useMutation({
    mutationFn: () => (snapshot.server.running ? api.stopServer() : api.startServer()),
    onSuccess: () => {
      setError("")
      void refresh()
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const onCopy = async (key: string, value: string) => {
    const ok = await copyText(value)
    if (!ok) return
    setCopiedKey(key)
    window.setTimeout(() => setCopiedKey(""), 1600)
  }

  return (
    <div className="space-y-3 pb-2">
      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="flex items-center gap-1.5 text-[12px] font-bold tracking-tight">
              <Waves className="size-3.5 text-[color:var(--color-signal)]" />
              Sortie WebRTC
            </h2>
            <p className="mt-0.5 text-[11px] text-muted-foreground">
              Seule sortie de lecture, latence minimale
            </p>
          </div>
          <Badge
            variant="outline"
            className={cn(
              "h-5 shrink-0 gap-1 border px-1.5 text-[8.5px] font-bold uppercase tracking-[0.08em]",
              stream.webrtcReady
                ? "border-[color:var(--color-signal)]/35 bg-[color:var(--color-volt)]/25 text-[#3f4a1a]"
                : "border-border text-muted-foreground",
            )}
          >
            <span
              className={cn(
                "size-1.5 rounded-full",
                stream.webrtcReady
                  ? "bg-[color:var(--color-live)] live-dot"
                  : "bg-muted-foreground/40",
              )}
            />
            {stream.webrtcReady ? "flux actif" : "hors ligne"}
          </Badge>
        </div>

        <div className="mt-2.5 rounded-xl bg-[color:var(--color-ink)] p-2.5">
          <p className="select-all break-all font-mono text-[10.5px] leading-relaxed text-[color:var(--color-volt)]">
            {webrtcUrl}
          </p>
          <p className="mt-1 font-mono text-[9px] uppercase tracking-[0.12em] text-[#8b8677]">
            ~0,2–0,5 s · port {stream.webrtcPort}
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
          <Button
            size="sm"
            variant="outline"
            className="h-9 shrink-0 bg-white"
            onClick={() => void onCopy("webrtc", webrtcUrl)}
          >
            {copiedKey === "webrtc" ? <Check /> : <Copy />}
            {copiedKey === "webrtc" ? "Copié" : "Copier"}
          </Button>
          <Button
            size="sm"
            className="h-9 shrink-0"
            onClick={() => openURL(webrtcUrl)}
          >
            <ExternalLink />
            Ouvrir
          </Button>
        </div>

        <div className="mt-2.5 flex items-start gap-1.5 text-[10px] leading-snug text-muted-foreground">
          <Volume2 className="mt-0.5 size-3 shrink-0 text-[color:var(--color-signal)]" />
          <span>
            {stream.webrtcAudio
              ? "Son de la caméra inclus : ffmpeg encode l'Opus directement avec l'image, en une seule passe."
              : stream.webrtcReady
                ? "Aucune piste audio sur la sortie : vérifiez que le son n'est pas coupé à la source."
                : stream.incoming.ready
                  ? incomingAudio
                    ? "Encodeur audio en cours de démarrage…"
                    : "La source n'envoie aucune piste audio."
                  : "En attente du flux source…"}
          </span>
        </div>

        <div className="mt-2.5 space-y-1.5">
          <PathRow label="Source → MediaMTX" path={stream.incoming} />
          <PathRow label="ffmpeg → WebRTC" path={stream.webrtc} />
        </div>

        {stream.encoderErr && (
          <p className="mt-2 rounded-lg bg-destructive/10 px-2 py-1.5 text-[10.5px] leading-snug text-destructive">
            {stream.encoderErr}
          </p>
        )}

        {!snapshot.environment.ffmpegReady && (
          <p className="mt-2 rounded-lg bg-[color:var(--color-warning)]/10 px-2 py-1.5 text-[10.5px] leading-snug text-[color:var(--color-warning)]">
            ffmpeg absent : installez-le (apt install ffmpeg). Sans lui, WebRTC ne peut
            pas recevoir l'audio Opus.
          </p>
        )}

        {stream.incoming.ready && !stream.webrtcReady && snapshot.environment.ffmpegReady && (
          <p className="mt-2 text-[10.5px] text-muted-foreground">
            ffmpeg prépare la copie WebRTC (H.264 + Opus)…
          </p>
        )}
      </section>

      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="flex items-center gap-1.5 text-[12px] font-bold tracking-tight">
              <Radio className="size-3.5 text-[color:var(--color-signal)]" />
              Source RTMP
            </h2>
            <p className="mt-0.5 text-[11px] text-muted-foreground">
              Entrée à saisir dans DJI Relay ou sur la caméra
            </p>
          </div>
          <Badge variant="outline" className="h-5 shrink-0 px-1.5 font-mono text-[8.5px]">
            :{stream.rtmpPort}
          </Badge>
        </div>

        <div className="mt-2.5 rounded-xl bg-[color:var(--color-ink)] p-2.5">
          <p className="select-all break-all font-mono text-[10.5px] leading-relaxed text-[color:var(--color-volt)]">
            {stream.directURL}
          </p>
        </div>
        <div className="mt-2 flex items-center gap-1.5">
          <Button
            size="sm"
            variant="outline"
            className="h-9 flex-1 bg-white"
            onClick={() => void onCopy("ingest", stream.directURL)}
          >
            {copiedKey === "ingest" ? <Check /> : <Copy />}
            {copiedKey === "ingest" ? "Copié" : "Copier l'URL d'entrée"}
          </Button>
        </div>
        <p className="mt-2 text-[10px] leading-snug text-muted-foreground">
          La caméra (ou le téléphone) pousse vers cette URL. WebRTC est produite à partir
          de ce flux : RTMP ne sert plus de sortie de lecture.
        </p>
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
            <span className="truncate text-[12px] font-bold tracking-tight">
              Serveur MediaMTX
            </span>
            <Badge variant="outline" className="h-4 shrink-0 px-1 font-mono text-[8.5px]">
              RTMP :{snapshot.server.rtmpPort}
            </Badge>
            <Badge variant="outline" className="h-4 shrink-0 px-1 font-mono text-[8.5px]">
              WebRTC :{stream.webrtcPort}
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
            ? "Le serveur écoute l'entrée RTMP et publie la sortie WebRTC."
            : "Démarrez le serveur avant de connecter la caméra ou le relais."}
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
