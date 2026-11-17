import { useEffect, useMemo, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  Bluetooth,
  BookmarkPlus,
  Check,
  Eye,
  EyeOff,
  Loader2,
  RadioTower,
  Send,
  ShieldAlert,
  SignalHigh,
  SignalLow,
  SignalMedium,
} from "lucide-react"
import { useForm } from "react-hook-form"
import { Badge } from "~/components/ui/badge"
import { Button } from "~/components/ui/button"
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

// BlueZ donne comme alias l'adresse MAC quand l'appareil n'a pas de nom :
// ces entrées sont du bruit, on ne les affiche qu'à la demande.
function isNamed(name: string): boolean {
  const value = name.trim()
  if (!value) return false
  return !/^([0-9a-f]{2}[-:]){5}[0-9a-f]{2}$/i.test(value)
}

function RssiIcon({ rssi }: { rssi: number }) {
  const cls = "size-3.5 shrink-0"
  if (rssi >= -60) return <SignalHigh className={cn(cls, "text-[color:var(--color-signal)]")} />
  if (rssi >= -75) return <SignalMedium className={cn(cls, "text-[color:var(--color-warning)]")} />
  if (rssi === 0) return <SignalLow className={cn(cls, "text-muted-foreground/35")} />
  return <SignalLow className={cn(cls, "text-muted-foreground/70")} />
}

function DeviceRow({
  device,
  selected,
  onSelect,
}: {
  device: Device
  selected: boolean
  onSelect: () => void
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      className={cn(
        "flex w-full items-center gap-2.5 rounded-xl border px-2.5 py-2 text-left transition-all",
        selected
          ? "border-[color:var(--color-signal)]/45 bg-[color:var(--color-volt)]/20"
          : "border-border bg-card hover:border-[color:var(--color-signal)]/30 hover:bg-muted/50",
      )}
    >
      <div
        className={cn(
          "grid size-7 shrink-0 place-items-center rounded-lg",
          device.isDJI
            ? "bg-[color:var(--color-volt)]/45 text-[#3f4a1a]"
            : "bg-muted text-muted-foreground",
        )}
      >
        <Bluetooth className="size-3.5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-[12px] font-semibold leading-tight">
            {device.name || "Appareil sans nom"}
          </span>
          {device.isDJI && (
            <Badge
              variant="outline"
              className="h-4 shrink-0 border-[color:var(--color-signal)]/35 bg-white px-1 text-[8.5px] font-bold tracking-wide text-[color:var(--color-signal)]"
            >
              DJI
            </Badge>
          )}
          {device.connected && (
            <Badge variant="secondary" className="h-4 shrink-0 px-1 text-[8.5px]">
              lié
            </Badge>
          )}
        </div>
        <p className="truncate font-mono text-[9.5px] text-muted-foreground">
          {device.address}
          {device.rssi !== 0 ? ` · ${device.rssi} dBm` : ""}
        </p>
      </div>
      <RssiIcon rssi={device.rssi} />
      <Check
        className={cn(
          "size-3.5 shrink-0 transition-opacity",
          selected ? "text-[color:var(--color-signal)]" : "opacity-0",
        )}
      />
    </button>
  )
}

function RadioCard({ snapshot }: { snapshot: Snapshot }) {
  const queryClient = useQueryClient()
  const [error, setError] = useState("")
  const radio = snapshot.radio

  const unblock = useMutation({
    mutationFn: api.unblockBluetooth,
    onSuccess: () => {
      setError("")
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  if (!radio.present || (!radio.soft && !radio.hard)) return null

  return (
    <div className="rounded-2xl border border-[color:var(--color-warning)]/25 bg-[#fff8e8] p-3">
      <div className="flex items-start gap-2">
        <ShieldAlert className="mt-px size-3.5 shrink-0 text-[color:var(--color-warning)]" />
        <div className="min-w-0 flex-1">
          <p className="text-[12px] font-semibold leading-snug">Radio Bluetooth coupée</p>
          <p className="mt-0.5 text-[11px] leading-snug text-[#6b5a2a]">
            {radio.hard
              ? "Un interrupteur physique (ou le BIOS) coupe la radio : rallumez-le pour scanner."
              : "Le noyau bloque la radio (rfkill) : aucun appareil ne peut apparaître tant qu'il est actif."}
          </p>
        </div>
      </div>
      {!radio.hard && (
        <Button
          size="sm"
          className="mt-2.5 w-full"
          onClick={() => unblock.mutate()}
          disabled={unblock.isPending}
        >
          {unblock.isPending ? <Loader2 className="animate-spin" /> : <RadioTower />}
          Débloquer le Bluetooth
        </Button>
      )}
      {(error || unblock.isError) && (
        <p className="mt-2 rounded-lg bg-destructive/10 px-2 py-1.5 text-[10.5px] leading-snug text-destructive">
          {error || errorMessage(unblock.error)}
        </p>
      )}
      {!radio.hard && (
        <p className="mt-2 text-[10px] leading-snug text-[#7a6a45]">
          Si le bouton échoue, une seule commande suffit :{" "}
          <span className="font-mono text-[#17150f]">sudo rfkill unblock bluetooth</span>
        </p>
      )}
    </div>
  )
}

type FormValues = { ssid: string; psk: string }

function ConnectForm({ device, snapshot }: { device: Device; snapshot: Snapshot }) {
  const queryClient = useQueryClient()
  const [showPsk, setShowPsk] = useState(false)
  const [error, setError] = useState("")
  const [savedNotice, setSavedNotice] = useState("")
  const connecting = snapshot.session.state === "starting"

  const saved = snapshot.settings.networks ?? []
  const primary = saved[0]
  const detected = snapshot.environment.wifiSSID

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

  const ssid = watch("ssid") ?? ""
  const psk = watch("psk") ?? ""
  const selectedProfile = saved.some((n) => n.ssid === ssid.trim()) ? ssid.trim() : "__new"

  useEffect(() => {
    const net = snapshot.settings.networks?.[0]
    reset({ ssid: net?.ssid || detected, psk: net?.psk || "" })
    setSavedNotice("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [device.address])

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
    setError("")
    setSavedNotice("")
    try {
      await api.startConnect({
        address: device.address,
        name: device.name,
        ssid: values.ssid.trim(),
        psk: values.psk,
      })
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
    } catch (err) {
      setError(errorMessage(err))
    }
  })

  return (
    <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h2 className="text-[12px] font-bold tracking-tight">Connexion forcée</h2>
          <p className="mt-0.5 text-[11px] leading-snug text-muted-foreground">
            Vers{" "}
            <span className="font-mono text-foreground">
              {snapshot.environment.localIP}:{snapshot.settings.rtmpPort}
            </span>
          </p>
        </div>
        <Badge
          variant="outline"
          className="h-5 shrink-0 border-[color:var(--color-signal)]/30 bg-[color:var(--color-volt)]/20 px-1.5 text-[9px] font-bold text-[#3f4a1a]"
        >
          {device.isDJI ? "DJI" : "BLE"}
        </Badge>
      </div>

      <form onSubmit={onSubmit} className="mt-3 space-y-3">
        <div className="space-y-1.5">
          <Label className="eyebrow block">Réseau mémorisé</Label>
          <div className="flex items-center gap-1.5">
            <Select value={selectedProfile} onValueChange={pickProfile}>
              <SelectTrigger size="sm" className="h-9 w-full min-w-0 flex-1 bg-white">
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
          <Label htmlFor="ssid" className="eyebrow block">
            SSID Wi-Fi <span className="text-destructive">*</span>
          </Label>
          <Input
            id="ssid"
            placeholder="MonReseau"
            autoComplete="off"
            {...register("ssid", {
              required: "Le SSID est obligatoire",
              minLength: { value: 1, message: "Le SSID est obligatoire" },
            })}
          />
          {errors.ssid && (
            <p className="text-[10.5px] text-destructive">{errors.ssid.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="psk" className="eyebrow block">
            Mot de passe
          </Label>
          <div className="relative">
            <Input
              id="psk"
              type={showPsk ? "text" : "password"}
              placeholder="••••••••"
              autoComplete="off"
              className="pr-10"
              {...register("psk")}
            />
            <button
              type="button"
              onClick={() => setShowPsk((v) => !v)}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
              aria-label={showPsk ? "Masquer" : "Afficher"}
            >
              {showPsk ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
            </button>
          </div>
          <p className="text-[10px] leading-snug text-muted-foreground">
            Vide = réseau ouvert. Le DJI exige un Wi-Fi 2,4 GHz.
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

        <Button type="submit" size="lg" className="w-full" disabled={connecting || isSubmitting}>
          {connecting || isSubmitting ? <Loader2 className="animate-spin" /> : <Send />}
          {connecting ? "Connexion en cours…" : "Forcer la connexion & diffuser"}
        </Button>
      </form>

      {snapshot.session.command && (
        <details className="mt-3 border-t border-border pt-2.5">
          <summary className="cursor-pointer list-none eyebrow hover:text-foreground">
            Commande djictl
          </summary>
          <pre className="mt-1.5 max-h-24 overflow-auto rounded-lg bg-[color:var(--color-ink)] p-2 font-mono text-[9px] leading-relaxed text-[#d9d4c5]">
            {snapshot.session.command}
          </pre>
        </details>
      )}
    </section>
  )
}

export function DevicesPanel({ snapshot }: { snapshot: Snapshot }) {
  const queryClient = useQueryClient()
  const [selected, setSelected] = useState<string>(snapshot.settings.lastDevice)
  const [showAll, setShowAll] = useState(false)
  const [error, setError] = useState("")

  const scanMutation = useMutation({
    mutationFn: () => (snapshot.scan.scanning ? api.stopScan() : api.startScan()),
    onSuccess: () => {
      setError("")
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const devices = useMemo(() => {
    const visible = snapshot.devices.filter((d) => showAll || d.isDJI || isNamed(d.name))
    return visible.sort((a, b) => {
      if (a.isDJI !== b.isDJI) return a.isDJI ? -1 : 1
      return 0
    })
  }, [snapshot.devices, showAll])

  const device = snapshot.devices.find((d) => d.address === selected) ?? null
  const djiCount = snapshot.devices.filter((d) => d.isDJI).length

  return (
    <div className="space-y-3 pb-2">
      <RadioCard snapshot={snapshot} />

      <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="text-[12px] font-bold tracking-tight">Appareils à portée</h2>
            <p className="mt-0.5 truncate text-[11px] text-muted-foreground">
              {snapshot.scan.scanning ? "scan en cours" : `${devices.length} détecté(s)`}
              {djiCount > 0 ? ` · ${djiCount} DJI` : ""}
            </p>
          </div>
          <Button
            size="sm"
            variant={snapshot.scan.scanning ? "outline" : "default"}
            onClick={() => scanMutation.mutate()}
            disabled={scanMutation.isPending}
            className="shrink-0"
          >
            {scanMutation.isPending ? <Loader2 className="animate-spin" /> : <Bluetooth />}
            {snapshot.scan.scanning ? "Arrêter" : "Scanner"}
          </Button>
        </div>

        {(snapshot.scan.err || error) && (
          <p className="mt-2.5 rounded-lg bg-destructive/10 px-2 py-1.5 text-[11px] leading-snug text-destructive">
            {snapshot.scan.err || error}
          </p>
        )}

        <div className="mt-2.5 max-h-[38vh] space-y-1.5 overflow-y-auto overscroll-contain pr-0.5">
          {devices.length === 0 ? (
            <div className="rounded-xl border border-dashed border-border px-3 py-6 text-center">
              <Bluetooth className="mx-auto size-5 text-muted-foreground/40" />
              <p className="mt-1.5 text-[11px] leading-snug text-muted-foreground">
                {snapshot.scan.scanning
                  ? "Recherche d'appareils Bluetooth…"
                  : "Lancez un scan pour lister les appareils à portée."}
              </p>
            </div>
          ) : (
            devices.map((d) => (
              <DeviceRow
                key={d.address}
                device={d}
                selected={d.address === selected}
                onSelect={() => setSelected(d.address)}
              />
            ))
          )}
        </div>

        {snapshot.devices.length > 0 && (
          <button
            type="button"
            onClick={() => setShowAll((v) => !v)}
            className="mt-2.5 text-[10px] text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
          >
            {showAll ? "Masquer les appareils sans nom" : "Afficher tous les appareils"}
          </button>
        )}
      </section>

      {device ? (
        <ConnectForm device={device} snapshot={snapshot} />
      ) : (
        <p className="px-1 text-[11px] leading-snug text-muted-foreground">
          Sélectionnez la caméra DJI dans la liste pour la forcer à rejoindre votre réseau.
        </p>
      )}
    </div>
  )
}
