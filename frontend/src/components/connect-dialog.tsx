import { useEffect, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { BookmarkPlus, Eye, EyeOff, Loader2, Send, Wifi } from "lucide-react"
import { useForm } from "react-hook-form"
import { Badge } from "~/components/ui/badge"
import { Button } from "~/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog"
import { Input } from "~/components/ui/input"
import { Label } from "~/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select"
import { api, errorMessage } from "~/lib/api"
import { cn } from "~/lib/utils"
import type { Device, Snapshot } from "~/lib/types"

type FormValues = { ssid: string; psk: string }

function RssiLabel({ rssi }: { rssi: number }) {
  if (rssi === 0) return null
  return <span>· {rssi} dBm</span>
}

/**
 * Formulaire de forçage BLE → Wi-Fi → RTMP, ouvert au clic sur un appareil.
 * `device` est l'instantané pris au moment du clic : la caméra peut disparaître
 * du scan dès qu'elle basculé sur le Wi-Fi, le dialogue doit rester utilisable.
 */
export function ConnectDialog({
  device,
  snapshot,
  open,
  onOpenChange,
}: {
  device: Device | null
  snapshot: Snapshot
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const queryClient = useQueryClient()
  const [showPsk, setShowPsk] = useState(false)
  const [error, setError] = useState("")
  const [savedNotice, setSavedNotice] = useState("")
  const [resolution, setResolution] = useState<"480p" | "720p" | "1080p">("1080p")
  const [bitrate, setBitrate] = useState("6000")
  const [fps, setFps] = useState<"25" | "30">("30")

  const saved = snapshot.settings.networks ?? []
  const primary = saved[0]
  const detected = snapshot.environment.wifiSSID
  const connecting = snapshot.session.state === "starting"

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    defaultValues: {
      ssid: primary?.ssid || detected,
      psk: primary?.psk || "",
    },
  })

  useEffect(() => {
    if (!open) return
    const net = saved[0]
    reset({ ssid: net?.ssid || detected, psk: net?.psk || "" })
    setResolution(snapshot.settings.streamResolution || "1080p")
    setBitrate(String(snapshot.settings.streamBitrate || 6000))
    setFps(String(snapshot.settings.streamFps || 30) as "25" | "30")
    setError("")
    setSavedNotice("")
    setShowPsk(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, device?.address])

  const ssid = watch("ssid") ?? ""
  const psk = watch("psk") ?? ""
  const selectedProfile = saved.some((n) => n.ssid === ssid.trim()) ? ssid.trim() : "__new"

  const pickProfile = (value: string) => {
    setSavedNotice("")
    if (value === "__new") {
      setValue("ssid", detected || "")
      setValue("psk", "")
      return
    }
    const net = saved.find((n) => n.ssid === value)
    if (net) {
      setValue("ssid", net.ssid)
      setValue("psk", net.psk)
    }
  }

  const remember = async () => {
    if (!ssid.trim()) return
    setError("")
    try {
      await api.saveNetwork(ssid.trim(), psk)
      setSavedNotice("Réseau mémorisé")
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  const onSubmit = handleSubmit(async (values) => {
    if (!device) return
    setError("")
    setSavedNotice("")
    try {
      await api.startConnect({
        address: device.address,
        name: device.name,
        ssid: values.ssid.trim(),
        psk: values.psk,
        resolution,
        bitrate: Number(bitrate),
        fps: Number(fps) as 25 | 30,
      })
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
      onOpenChange(false)
    } catch (err) {
      setError(errorMessage(err))
    }
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88dvh] gap-3 overflow-y-auto rounded-2xl p-3.5">
        <DialogHeader className="gap-1.5">
          <p className="eyebrow">Connexion forcée</p>
          <div className="flex items-center gap-2 pr-7">
            <DialogTitle className="min-w-0 truncate text-[14px] leading-tight">
              {device?.name || "Appareil sans nom"}
            </DialogTitle>
            <Badge
              variant="outline"
              className="h-4 shrink-0 border-[color:var(--color-signal)]/35 bg-[color:var(--color-volt)]/25 px-1 text-[8.5px] font-bold tracking-wide text-[#3f4a1a]"
            >
              {device?.isDJI ? "DJI" : "BLE"}
            </Badge>
          </div>
          <DialogDescription className="flex items-center gap-1.5 font-mono text-[9.5px]">
            <span className="truncate">{device?.address}</span>
            {device && <RssiLabel rssi={device.rssi} />}
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 rounded-xl bg-[color:var(--color-ink)] px-2.5 py-2">
          <Wifi className="size-3 shrink-0 text-[color:var(--color-volt)]" />
          <span className="eyebrow shrink-0 text-[#8b8677]">cible</span>
          <span className="min-w-0 flex-1 truncate font-mono text-[9.5px] text-[color:var(--color-volt)]">
            rtmp://{snapshot.environment.localIP}:{snapshot.settings.rtmpPort}/live/dji
          </span>
        </div>

        <form onSubmit={onSubmit} className="space-y-3">
          <div className="space-y-1.5">
            <Label className="eyebrow block">Réseau mémorisé</Label>
            <div className="flex items-center gap-1.5">
              <Select value={selectedProfile} onValueChange={pickProfile}>
                <SelectTrigger className="h-9 w-full min-w-0 flex-1 bg-white">
                  <SelectValue placeholder="Choisir un réseau" />
                </SelectTrigger>
                <SelectContent className="max-w-[280px]">
                  {saved.map((net) => (
                    <SelectItem key={net.ssid} value={net.ssid}>
                      {net.ssid}
                    </SelectItem>
                  ))}
                  <SelectItem value="__new">Nouveau réseau…</SelectItem>
                </SelectContent>
              </Select>
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={remember}
                disabled={!ssid.trim()}
                aria-label="Mémoriser ce réseau"
                title="Mémoriser ce réseau"
                className="shrink-0 bg-white"
              >
                <BookmarkPlus />
              </Button>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="connect-ssid" className="eyebrow block">
              SSID Wi-Fi <span className="text-destructive">*</span>
            </Label>
            <Input
              id="connect-ssid"
              placeholder="MonReseau"
              autoComplete="off"
              {...register("ssid", {
                required: "Le SSID est obligatoire",
                minLength: { value: 1, message: "Le SSID est obligatoire" },
              })}
            />
            {errors.ssid && <p className="text-[10.5px] text-destructive">{errors.ssid.message}</p>}
            {detected && detected !== ssid.trim() && (
              <button
                type="button"
                onClick={() => setValue("ssid", detected)}
                className="text-[10px] text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
              >
                Utiliser le réseau détecté · {detected}
              </button>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="connect-psk" className="eyebrow block">
              Mot de passe
            </Label>
            <div className="relative">
              <Input
                id="connect-psk"
                type={showPsk ? "text" : "password"}
                placeholder="Mot de passe du Wi-Fi"
                autoComplete="off"
                className="pr-10"
                {...register("psk")}
              />
              <button
                type="button"
                onClick={() => setShowPsk((v) => !v)}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                aria-label={showPsk ? "Masquer le mot de passe" : "Afficher le mot de passe"}
              >
                {showPsk ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
              </button>
            </div>
            <p className="text-[10px] leading-snug text-muted-foreground">
              Vide = réseau ouvert. La caméra exige un Wi-Fi 2,4 GHz (WPA2).
            </p>
          </div>

          <div className="space-y-1.5">
            <Label className="eyebrow block">Encodeur de la caméra</Label>
            <div className="grid grid-cols-[1.1fr_1fr_0.8fr] gap-2">
              <Select value={resolution} onValueChange={(v) => setResolution(v as "480p" | "720p" | "1080p")}>
                <SelectTrigger className="bg-white" aria-label="Résolution">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1080p">1080p</SelectItem>
                  <SelectItem value="720p">720p</SelectItem>
                  <SelectItem value="480p">480p</SelectItem>
                </SelectContent>
              </Select>
              <Select value={bitrate} onValueChange={setBitrate}>
                <SelectTrigger className="bg-white" aria-label="Débit">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="4000">4 Mb/s</SelectItem>
                  <SelectItem value="6000">6 Mb/s</SelectItem>
                  <SelectItem value="8000">8 Mb/s</SelectItem>
                  <SelectItem value="12000">12 Mb/s</SelectItem>
                  <SelectItem value="20000">20 Mb/s</SelectItem>
                </SelectContent>
              </Select>
              <Select value={fps} onValueChange={(v) => setFps(v as "25" | "30")}>
                <SelectTrigger className="bg-white" aria-label="Images par seconde">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="30">30 fps</SelectItem>
                  <SelectItem value="25">25 fps</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <p className="text-[10px] leading-snug text-muted-foreground">
              Appliqués à distance par Bluetooth : rien à régler dans le menu du DJI.
            </p>
          </div>

          {error && (
            <p className="rounded-lg bg-destructive/10 px-2 py-1.5 text-[11px] leading-snug text-destructive">
              {error}
            </p>
          )}
          {savedNotice && (
            <p className="rounded-lg bg-[color:var(--color-volt)]/30 px-2 py-1.5 text-[11px] text-[#3f4a1a]">
              {savedNotice}
            </p>
          )}

          <div className={cn("-mx-3.5 -mb-3.5 flex flex-row gap-2 rounded-b-2xl border-t border-border bg-muted/40 p-3")}>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="flex-1 bg-white"
            >
              Annuler
            </Button>
            <Button type="submit" disabled={connecting || isSubmitting || !device} className="flex-[1.4]">
              {connecting || isSubmitting ? <Loader2 className="animate-spin" /> : <Send />}
              {connecting ? "Envoi…" : "Connecter"}
            </Button>
          </div>
        </form>

        {snapshot.session.command && (
          <details className="-mt-1 border-t border-border pt-2.5">
            <summary className="cursor-pointer list-none eyebrow hover:text-foreground">
              Commande djictl
            </summary>
            <pre className="mt-1.5 max-h-24 overflow-auto rounded-lg bg-[color:var(--color-ink)] p-2 font-mono text-[9px] leading-relaxed text-[#d9d4c5]">
              {snapshot.session.command}
            </pre>
          </details>
        )}
      </DialogContent>
    </Dialog>
  )
}
