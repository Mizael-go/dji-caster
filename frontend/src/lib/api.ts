import * as Backend from "../../wailsjs/go/main/App"
import { ClipboardSetText } from "../../wailsjs/runtime/runtime"
import type { ConnectRequest, LogLine, Settings, Snapshot } from "./types"

// Toutes les communications passent par le RPC généré par Wails : pas de fetch
// HTTP, pas d'axios.
export const api = {
  snapshot: () => Backend.Snapshot() as unknown as Promise<Snapshot>,
  logs: () => Backend.Logs() as unknown as Promise<LogLine[]>,
  clearLogs: () => Backend.ClearLogs() as Promise<void>,

  startScan: () => Backend.StartScan() as Promise<void>,
  stopScan: () => Backend.StopScan() as Promise<void>,
  unblockBluetooth: () => Backend.UnblockBluetooth() as Promise<void>,

  saveNetwork: (ssid: string, psk: string) =>
    Backend.SaveNetwork(ssid, psk) as Promise<void>,
  deleteNetwork: (ssid: string) => Backend.DeleteNetwork(ssid) as Promise<void>,

  startServer: () => Backend.StartServer() as Promise<void>,
  stopServer: () => Backend.StopServer() as Promise<void>,
  setTranscode: (enabled: boolean) => Backend.SetTranscode(enabled) as Promise<void>,

  startConnect: (req: ConnectRequest) => Backend.StartConnect(req as never) as Promise<void>,
  cancelConnect: () => Backend.CancelConnect() as Promise<void>,
  resetSession: () => Backend.ResetSession() as Promise<void>,

  saveSettings: (settings: Settings) => Backend.SaveSettings(settings as never) as Promise<void>,
  installTools: () => Backend.InstallTools() as Promise<void>,
  openLogsFolder: () => Backend.OpenLogsFolder() as Promise<void>,
}

export async function copyText(text: string): Promise<boolean> {
  try {
    return await ClipboardSetText(text)
  } catch {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      return false
    }
  }
}

export function errorMessage(err: unknown): string {
  if (!err) return ""
  if (typeof err === "string") return err
  if (err instanceof Error) return err.message
  return String(err)
}
