import { useEffect, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { AlertTriangle, Check, Download, FolderOpen, Loader2 } from "lucide-react"
import { useForm } from "react-hook-form"
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
import { Separator } from "~/components/ui/separator"
import { api, errorMessage } from "~/lib/api"
import { cn } from "~/lib/utils"
import type { Snapshot } from "~/lib/types"

type FormValues = {
  rtmpPort: number
  apiPort: number
}

function DependencyRow({
  label,
  ok,
  path,
  hint,
}: {
  label: string
  ok: boolean
  path: string
  hint: string
}) {
  return (
    <div className="flex items-start gap-2">
      {ok ? (
        <Check className="mt-0.5 size-3 shrink-0 text-[color:var(--color-signal)]" />
      ) : (
        <AlertTriangle className="mt-0.5 size-3 shrink-0 text-[color:var(--color-warning)]" />
      )}
      <div className="min-w-0 flex-1">
        <p className="text-[11px] font-medium">{label}</p>
        <p
          className={cn(
            "truncate font-mono text-[9px]",
            ok ? "text-muted-foreground" : "text-[color:var(--color-warning)]",
          )}
          title={path || hint}
        >
          {ok ? path : hint}
        </p>
      </div>
    </div>
  )
}

export function SettingsDialog({
  snapshot,
  open,
  onOpenChange,
}: {
  snapshot: Snapshot
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const queryClient = useQueryClient()
  const [error, setError] = useState("")
  const env = snapshot.environment

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    defaultValues: {
      rtmpPort: snapshot.settings.rtmpPort,
      apiPort: snapshot.settings.apiPort,
    },
  })

  useEffect(() => {
    if (open) {
      reset({
        rtmpPort: snapshot.settings.rtmpPort,
        apiPort: snapshot.settings.apiPort,
      })
      setError("")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const saveMutation = useMutation({
    mutationFn: async (values: FormValues) => {
      await api.saveSettings({
        ...snapshot.settings,
        rtmpPort: Number(values.rtmpPort),
        apiPort: Number(values.apiPort),
      })
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
      onOpenChange(false)
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const installMutation = useMutation({
    mutationFn: api.installTools,
    onSuccess: () => {
      setError("")
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const allReady = env.mediaMtxReady && env.ffmpegReady

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88dvh] gap-3 overflow-y-auto rounded-2xl p-3.5">
        <DialogHeader>
          <DialogTitle className="text-[14px]">Réglages</DialogTitle>
          <DialogDescription className="text-[11px] leading-snug">
            Ports du serveur et outils locaux.
          </DialogDescription>
        </DialogHeader>

        <form
          onSubmit={handleSubmit((values) => saveMutation.mutate(values))}
          className="space-y-3"
        >
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <Label htmlFor="set-rtmp" className="eyebrow block">
                Port RTMP
              </Label>
              <Input
                id="set-rtmp"
                type="number"
                inputMode="numeric"
                {...register("rtmpPort", {
                  valueAsNumber: true,
                  min: { value: 1, message: "1–65535" },
                  max: { value: 65535, message: "1–65535" },
                })}
              />
              {errors.rtmpPort && (
                <p className="text-[10px] text-destructive">{errors.rtmpPort.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="set-api" className="eyebrow block">
                Port API
              </Label>
              <Input
                id="set-api"
                type="number"
                inputMode="numeric"
                {...register("apiPort", {
                  valueAsNumber: true,
                  min: { value: 1, message: "1–65535" },
                  max: { value: 65535, message: "1–65535" },
                })}
              />
              {errors.apiPort && (
                <p className="text-[10px] text-destructive">{errors.apiPort.message}</p>
              )}
            </div>
          </div>

          <Separator />

          <div className="space-y-2">
            <p className="eyebrow">Outils</p>
            <DependencyRow
              label="MediaMTX (copie dédiée à l'app)"
              ok={env.mediaMtxReady}
              path={env.mediaMtxPath}
              hint="non installé"
            />
            <DependencyRow
              label="ffmpeg (encodage WebRTC H.264 + Opus)"
              ok={env.ffmpegReady}
              path={env.ffmpegPath}
              hint="absent du PATH"
            />
            <div className="flex items-center gap-1.5 pt-0.5">
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => installMutation.mutate()}
                disabled={installMutation.isPending}
                className="flex-1 bg-white"
              >
                {installMutation.isPending ? (
                  <Loader2 className="animate-spin" />
                ) : (
                  <Download />
                )}
                {allReady ? "Mettre à jour MediaMTX" : "Installer MediaMTX"}
              </Button>
              <Button
                type="button"
                size="icon-sm"
                variant="outline"
                onClick={() => void api.openLogsFolder()}
                aria-label="Ouvrir le dossier de données"
                className="shrink-0 bg-white"
              >
                <FolderOpen />
              </Button>
            </div>
            <p
              className="truncate font-mono text-[9px] text-muted-foreground"
              title={env.dataDir}
            >
              {env.dataDir}
            </p>
          </div>

          {error && (
            <p className="rounded-lg bg-destructive/10 px-2 py-1.5 text-[10.5px] leading-snug text-destructive">
              {error}
            </p>
          )}

          <div className="-mx-3.5 -mb-3.5 flex flex-row gap-2 rounded-b-2xl border-t border-border bg-muted/40 p-3">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="flex-1 bg-white"
            >
              Fermer
            </Button>
            <Button type="submit" disabled={saveMutation.isPending} className="flex-1">
              {saveMutation.isPending && <Loader2 className="animate-spin" />}
              Enregistrer
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
