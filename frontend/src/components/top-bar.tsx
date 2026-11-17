import { Settings as SettingsIcon } from "lucide-react"
import { Button } from "~/components/ui/button"
import { cn } from "~/lib/utils"
import type { SessionState, Snapshot } from "~/lib/types"

export function BrandMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 40 40" className={className} aria-hidden="true">
      <rect width="40" height="40" rx="11" fill="#17150f" />
      <circle cx="20" cy="26.4" r="1.9" fill="#d8ff3e" />
      <g fill="none" stroke="#d8ff3e" strokeWidth="2.4" strokeLinecap="round">
        <path d="M14.6 24.9a7.6 7.6 0 0 1 0-9.8" />
        <path d="M25.4 15.1a7.6 7.6 0 0 1 0 9.8" />
        <path d="M10.6 29a13.6 13.6 0 0 1 0-18" />
        <path d="M29.4 11a13.6 13.6 0 0 1 0 18" />
      </g>
    </svg>
  )
}

const STATE_STYLES: Record<
  SessionState,
  { label: string; className: string; dot: string }
> = {
  idle: {
    label: "PRÊT",
    className: "text-muted-foreground",
    dot: "bg-muted-foreground/40",
  },
  starting: {
    label: "CONNEXION",
    className: "text-[color:var(--color-warning)]",
    dot: "bg-[color:var(--color-warning)] animate-pulse",
  },
  waiting: {
    label: "EN ATTENTE",
    className: "text-[color:var(--color-warning)]",
    dot: "bg-[color:var(--color-warning)] animate-pulse",
  },
  live: {
    label: "DIRECT",
    className: "text-[color:var(--color-live)]",
    dot: "bg-[color:var(--color-live)] live-dot",
  },
  error: {
    label: "ERREUR",
    className: "text-destructive",
    dot: "bg-destructive",
  },
}

export function TopBar({
  snapshot,
  onOpenSettings,
}: {
  snapshot: Snapshot
  onOpenSettings: () => void
}) {
  const state = STATE_STYLES[snapshot.session.state] ?? STATE_STYLES.idle
  const subtitle = snapshot.server.running
    ? `${snapshot.environment.localIP}:${snapshot.server.rtmpPort}`
    : "serveur arrêté"

  return (
    <header className="safe-top z-20 shrink-0 border-b border-border/70 bg-background/85 backdrop-blur-md">
      <div className="flex items-center gap-2.5 px-4 py-2.5">
        <BrandMark className="size-8 shrink-0" />
        <div className="min-w-0 flex-1">
          <h1 className="truncate text-[14px] font-extrabold leading-tight tracking-tight">
            DJI Caster
          </h1>
          <p className="truncate font-mono text-[9.5px] uppercase tracking-[0.14em] text-muted-foreground">
            {subtitle}
          </p>
        </div>
        <span
          className={cn(
            "flex shrink-0 items-center gap-1.5 text-[9px] font-bold tracking-[0.12em]",
            state.className,
          )}
        >
          <span className={cn("size-1.5 rounded-full", state.dot)} />
          {state.label}
        </span>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onOpenSettings}
          aria-label="Réglages"
          className="-mr-1 text-muted-foreground"
        >
          <SettingsIcon />
        </Button>
      </div>
    </header>
  )
}
