// Miroir TypeScript des structures JSON exposées par le backend Go.

export type Network = {
  ssid: string
  psk: string
  lastUsed: number
}

export type Settings = {
  networks: Network[]
  rtmpPort: number
  apiPort: number
  transcode: boolean
  lastDevice: string
  mediaMtxPath: string
  djiCtlPath: string
  ffmpegPath: string
}

export type Device = {
  address: string
  name: string
  rssi: number
  connected: boolean
  paired: boolean
  isDJI: boolean
  lastSeen: number
}

export type EnvView = {
  os: string
  arch: string
  localIP: string
  wifiSSID: string
  dataDir: string
  mediaMtxPath: string
  djiCtlPath: string
  ffmpegPath: string
  mediaMtxReady: boolean
  djiCtlReady: boolean
  ffmpegReady: boolean
}

export type ScanView = {
  scanning: boolean
  err: string
}

// Interrupteur radio du noyau (rfkill).
export type Radio = {
  present: boolean
  soft: boolean
  hard: boolean
}

export type ServerView = {
  running: boolean
  err: string
  rtmpPort: number
  apiPort: number
}

export type SessionState = "idle" | "starting" | "waiting" | "live" | "error"

export type SessionView = {
  state: SessionState
  err: string
  hint: string
  deviceAddr: string
  deviceName: string
  command: string
}

export type PathView = {
  ready: boolean
  codecs: string[]
  bytes: number
  readers: number
}

export type StreamView = {
  host: string
  rtmpPort: number
  directURL: string
  obsURL: string
  localURL: string
  transcode: boolean
  transcodeErr: string
  transcoding: boolean
  incoming: PathView
  outgoing: PathView
}

export type Snapshot = {
  settings: Settings
  environment: EnvView
  radio: Radio
  scan: ScanView
  devices: Device[]
  server: ServerView
  session: SessionView
  stream: StreamView
}

export type LogLine = {
  scope: string
  level: string
  text: string
  time: number
}

export type ConnectRequest = {
  address: string
  name: string
  ssid: string
  psk: string
}
