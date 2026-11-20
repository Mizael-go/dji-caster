// Miroir TypeScript des structures JSON exposées par le backend Go.

export type Settings = {
  rtmpPort: number
  apiPort: number
  mediaMtxPath: string
  ffmpegPath: string
}

export type EnvView = {
  os: string
  arch: string
  localIP: string
  dataDir: string
  mediaMtxPath: string
  ffmpegPath: string
  mediaMtxReady: boolean
  ffmpegReady: boolean
}

export type ServerView = {
  running: boolean
  err: string
  rtmpPort: number
  apiPort: number
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
  webrtcPort: number
  webrtcURL: string
  webrtcLocalURL: string
  webrtcReady: boolean
  webrtcAudio: boolean
  encoderErr: string
  incoming: PathView
  webrtc: PathView
}

export type Snapshot = {
  settings: Settings
  environment: EnvView
  server: ServerView
  stream: StreamView
}

export type LogLine = {
  scope: string
  level: string
  text: string
  time: number
}
