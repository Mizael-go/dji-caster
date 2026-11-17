import { useEffect } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { EventsOn } from "../../wailsjs/runtime/runtime"
import { api } from "~/lib/api"
import type { LogLine, Snapshot } from "~/lib/types"

export function useSnapshot() {
  return useQuery<Snapshot>({
    queryKey: ["snapshot"],
    queryFn: api.snapshot,
    refetchInterval: 1000,
    staleTime: 0,
    retry: false,
  })
}

export function useLogs() {
  return useQuery<LogLine[]>({
    queryKey: ["logs"],
    queryFn: api.logs,
    refetchInterval: 1200,
    retry: false,
  })
}

// Le backend pousse un événement dès qu'un état change (scan, session, serveur),
// ce qui rafraîchit immédiatement l'UI sans attendre le prochain poll.
export function useBackendEvents() {
  const queryClient = useQueryClient()
  useEffect(() => {
    const off = EventsOn("state", () => {
      void queryClient.invalidateQueries({ queryKey: ["snapshot"] })
      void queryClient.invalidateQueries({ queryKey: ["logs"] })
    })
    return () => {
      if (typeof off === "function") off()
    }
  }, [queryClient])
}
