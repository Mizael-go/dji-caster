import { useState } from "react"
import { Bluetooth, Loader2, Radio, Terminal } from "lucide-react"
import { DevicesPanel } from "~/components/devices-panel"
import { LogsPanel } from "~/components/logs-panel"
import { SessionBanner } from "~/components/session-banner"
import { SettingsDialog } from "~/components/settings-dialog"
import { StreamPanel } from "~/components/stream-panel"
import { BrandMark, TopBar } from "~/components/top-bar"
import { Button } from "~/components/ui/button"
import { Skeleton } from "~/components/ui/skeleton"
import { useBackendEvents, useLogs, useSnapshot } from "~/hooks/use-app"
import { api, errorMessage } from "~/lib/api"
import { cn } from "~/lib/utils"

const TABS = [
  { id: "devices", label: "Appareils", icon: Bluetooth },
  { id: "stream", label: "Diffusion", icon: Radio },
  { id: "logs", label: "Journaux", icon: Terminal },
] as const

type TabId = (typeof TABS)[number]["id"]

export default function App() {
  const snapshotQuery = useSnapshot()
  const logsQuery = useLogs()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [tab, setTab] = useState<TabId>("devices")

  useBackendEvents()

  const snapshot = snapshotQuery.data

  if (!snapshot) {
    return (
      <div className="app-shell-bg flex h-dvh flex-col overflow-hidden">
        <div className="flex items-center gap-2.5 border-b border-border/70 px-4 py-2.5">
          <BrandMark className="size-8" />
          <div>
            <p className="text-[14px] font-extrabold leading-tight tracking-tight">DJI Caster</p>
            <p className="font-mono text-[9.5px] uppercase tracking-[0.14em] text-muted-foreground">
              démarrage…
            </p>
          </div>
        </div>
        <div className="space-y-3 p-4">
          <div className="flex items-center gap-2">
            <Loader2 className="size-3.5 animate-spin text-[color:var(--color-signal)]" />
            <span className="text-[11.5px] text-muted-foreground">Connexion au moteur local…</span>
          </div>
          <Skeleton className="h-24 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
          {snapshotQuery.error && (
            <div className="rounded-2xl border border-destructive/30 bg-destructive/[0.06] p-3">
              <p className="text-[11px] text-destructive">{errorMessage(snapshotQuery.error)}</p>
              <Button
                size="sm"
                variant="outline"
                className="mt-2"
                onClick={() => void snapshotQuery.refetch()}
              >
                Réessayer
              </Button>
            </div>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="app-shell-bg flex h-dvh flex-col overflow-hidden">
      <TopBar snapshot={snapshot} onOpenSettings={() => setSettingsOpen(true)} />

      <main className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pb-5">
        <SessionBanner session={snapshot.session} onReset={() => void api.resetSession()} />

        <div className="mt-3">
          {tab === "devices" && <DevicesPanel snapshot={snapshot} />}
          {tab === "stream" && <StreamPanel snapshot={snapshot} />}
          {tab === "logs" && <LogsPanel logs={logsQuery.data ?? []} />}
        </div>
      </main>

      <nav
        aria-label="Navigation principale"
        className="safe-bottom shrink-0 border-t border-border/70 bg-background/90 backdrop-blur-xl"
      >
        <div className="flex h-[58px] items-stretch px-3">
          {TABS.map((item) => {
            const active = tab === item.id
            return (
              <button
                key={item.id}
                type="button"
                onClick={() => setTab(item.id)}
                aria-current={active ? "page" : undefined}
                className={cn(
                  "relative flex flex-1 flex-col items-center justify-center gap-1 text-[10px] font-medium transition-colors",
                  active
                    ? "text-foreground"
                    : "text-muted-foreground/70 hover:text-muted-foreground active:text-foreground",
                )}
              >
                {active && (
                  <span className="absolute top-0 h-[2px] w-9 rounded-full bg-[color:var(--color-signal)]" />
                )}
                <item.icon className="size-[17px]" />
                {item.label}
              </button>
            )
          })}
        </div>
      </nav>

      <SettingsDialog
        snapshot={snapshot}
        open={settingsOpen}
        onOpenChange={setSettingsOpen}
      />
    </div>
  )
}
