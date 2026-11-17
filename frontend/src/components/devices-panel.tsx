import { useMemo, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  Bluetooth,
  ChevronRight,
  Loader2,
  RadioTower,
  ShieldAlert,
  SignalHigh,
  SignalLow,
  SignalMedium,
} from "lucide-react"
import { ConnectDialog } from "~/components/connect-dialog"
import { Badge } from "~/components/ui/badge"
import { Button } from "~/components/ui/button"
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

function DeviceRow({ device, onSelect }: { device: Device; onSelect: () => void }) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className="group flex w-full items-center gap-2.5 rounded-xl border border-border bg-card px-2.5 py-2 text-left transition-all hover:border-[color:var(--color-signal)]/40 hover:bg-[color:var(--color-volt)]/10 active:scale-[0.995]"
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
      <ChevronRight className="size-3.5 shrink-0 text-muted-foreground/50 transition-transform group-hover:translate-x-0.5 group-hover:text-[color:var(--color-signal)]" />
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

export function DevicesPanel({ snapshot }: { snapshot: Snapshot }) {
  const queryClient = useQueryClient()
  const [target, setTarget] = useState<Device | null>(null)
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

        <div className="mt-2.5 max-h-[46vh] space-y-1.5 overflow-y-auto overscroll-contain pr-0.5">
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
              <DeviceRow key={d.address} device={d} onSelect={() => setTarget(d)} />
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

      <p className="flex items-start gap-1.5 px-1 text-[11px] leading-snug text-muted-foreground">
        <span className="font-mono text-[9.5px] uppercase tracking-[0.14em] text-[color:var(--color-signal)]">
          astuce
        </span>
        <span>
          Touchez la caméra DJI pour la forcer à rejoindre votre Wi-Fi et démarrer le flux RTMP.
        </span>
      </p>

      <ConnectDialog
        device={target}
        snapshot={snapshot}
        open={target !== null}
        onOpenChange={(open) => {
          if (!open) setTarget(null)
        }}
      />
    </div>
  )
}
