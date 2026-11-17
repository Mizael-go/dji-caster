import { useEffect, useRef } from "react"
import { Eraser, Terminal } from "lucide-react"
import { Button } from "~/components/ui/button"
import { api } from "~/lib/api"
import { cn, clock } from "~/lib/utils"
import type { LogLine } from "~/lib/types"

const SCOPE_COLORS: Record<string, string> = {
  mediamtx: "text-[#8fd0ff]",
  ffmpeg: "text-[#ffcf6b]",
  djictl: "text-[color:var(--color-volt)]",
  ble: "text-[#c9a7ff]",
  app: "text-[#a8a396]",
}

export function LogsPanel({ logs }: { logs: LogLine[] }) {
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" })
  }, [logs.length])

  return (
    <section className="rounded-2xl border border-border bg-card p-3.5 card-raised">
      <div className="flex items-center justify-between gap-2">
        <h2 className="flex items-center gap-1.5 text-[12px] font-bold tracking-tight">
          <Terminal className="size-3.5 text-[color:var(--color-signal)]" />
          Console
        </h2>
        <Button
          size="xs"
          variant="ghost"
          onClick={() => void api.clearLogs()}
          disabled={logs.length === 0}
          className="-mr-1 text-muted-foreground"
        >
          <Eraser />
          Vider
        </Button>
      </div>

      <div className="mt-2.5 max-h-[calc(100dvh-260px)] min-h-[240px] overflow-y-auto rounded-xl bg-[color:var(--color-ink)] px-2.5 py-2 font-mono text-[9.5px] leading-relaxed">
        {logs.length === 0 ? (
          <p className="py-8 text-center text-[#7d786a]">Aucun journal pour l'instant.</p>
        ) : (
          <>
            {logs.map((line, i) => (
              <div key={`${line.time}-${i}`} className="flex gap-1.5 py-px">
                <span className="shrink-0 text-[#6f6a5c]">{clock(line.time)}</span>
                <span
                  className={cn(
                    "w-[52px] shrink-0",
                    SCOPE_COLORS[line.scope] ?? "text-[#a8a396]",
                  )}
                >
                  {line.scope}
                </span>
                <span
                  className={cn(
                    "min-w-0 flex-1 break-words",
                    line.level === "warn" ? "text-[#ffcf6b]" : "text-[#ddd8c9]",
                  )}
                >
                  {line.text}
                </span>
              </div>
            ))}
            <div ref={endRef} />
          </>
        )}
      </div>
    </section>
  )
}
