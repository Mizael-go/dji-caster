export namespace config {
	
	export class Network {
	    ssid: string;
	    psk: string;
	    lastUsed: number;
	
	    static createFrom(source: any = {}) {
	        return new Network(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ssid = source["ssid"];
	        this.psk = source["psk"];
	        this.lastUsed = source["lastUsed"];
	    }
	}
	export class Settings {
	    networks: Network[];
	    rtmpPort: number;
	    apiPort: number;
	    transcode: boolean;
	    lastDevice: string;
	    mediaMtxPath: string;
	    djiCtlPath: string;
	    ffmpegPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.networks = this.convertValues(source["networks"], Network);
	        this.rtmpPort = source["rtmpPort"];
	        this.apiPort = source["apiPort"];
	        this.transcode = source["transcode"];
	        this.lastDevice = source["lastDevice"];
	        this.mediaMtxPath = source["mediaMtxPath"];
	        this.djiCtlPath = source["djiCtlPath"];
	        this.ffmpegPath = source["ffmpegPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class ConnectRequest {
	    address: string;
	    name: string;
	    ssid: string;
	    psk: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.name = source["name"];
	        this.ssid = source["ssid"];
	        this.psk = source["psk"];
	    }
	}
	export class DeviceView {
	    address: string;
	    name: string;
	    rssi: number;
	    connected: boolean;
	    paired: boolean;
	    isDJI: boolean;
	    lastSeen: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.name = source["name"];
	        this.rssi = source["rssi"];
	        this.connected = source["connected"];
	        this.paired = source["paired"];
	        this.isDJI = source["isDJI"];
	        this.lastSeen = source["lastSeen"];
	    }
	}
	export class EnvView {
	    os: string;
	    arch: string;
	    localIP: string;
	    wifiSSID: string;
	    dataDir: string;
	    mediaMtxPath: string;
	    djiCtlPath: string;
	    ffmpegPath: string;
	    mediaMtxReady: boolean;
	    djiCtlReady: boolean;
	    ffmpegReady: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EnvView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.localIP = source["localIP"];
	        this.wifiSSID = source["wifiSSID"];
	        this.dataDir = source["dataDir"];
	        this.mediaMtxPath = source["mediaMtxPath"];
	        this.djiCtlPath = source["djiCtlPath"];
	        this.ffmpegPath = source["ffmpegPath"];
	        this.mediaMtxReady = source["mediaMtxReady"];
	        this.djiCtlReady = source["djiCtlReady"];
	        this.ffmpegReady = source["ffmpegReady"];
	    }
	}
	export class LogLine {
	    scope: string;
	    level: string;
	    text: string;
	    time: number;
	
	    static createFrom(source: any = {}) {
	        return new LogLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scope = source["scope"];
	        this.level = source["level"];
	        this.text = source["text"];
	        this.time = source["time"];
	    }
	}
	export class PathView {
	    ready: boolean;
	    codecs: string[];
	    bytes: number;
	    readers: number;
	
	    static createFrom(source: any = {}) {
	        return new PathView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.codecs = source["codecs"];
	        this.bytes = source["bytes"];
	        this.readers = source["readers"];
	    }
	}
	export class RadioView {
	    present: boolean;
	    soft: boolean;
	    hard: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RadioView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.present = source["present"];
	        this.soft = source["soft"];
	        this.hard = source["hard"];
	    }
	}
	export class ScanView {
	    scanning: boolean;
	    err: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scanning = source["scanning"];
	        this.err = source["err"];
	    }
	}
	export class ServerView {
	    running: boolean;
	    err: string;
	    rtmpPort: number;
	    apiPort: number;
	
	    static createFrom(source: any = {}) {
	        return new ServerView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.err = source["err"];
	        this.rtmpPort = source["rtmpPort"];
	        this.apiPort = source["apiPort"];
	    }
	}
	export class SessionView {
	    state: string;
	    err: string;
	    hint: string;
	    deviceAddr: string;
	    deviceName: string;
	    command: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.err = source["err"];
	        this.hint = source["hint"];
	        this.deviceAddr = source["deviceAddr"];
	        this.deviceName = source["deviceName"];
	        this.command = source["command"];
	    }
	}
	export class StreamView {
	    host: string;
	    rtmpPort: number;
	    directURL: string;
	    obsURL: string;
	    localURL: string;
	    transcode: boolean;
	    transcodeErr: string;
	    transcoding: boolean;
	    incoming: PathView;
	    outgoing: PathView;
	
	    static createFrom(source: any = {}) {
	        return new StreamView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.rtmpPort = source["rtmpPort"];
	        this.directURL = source["directURL"];
	        this.obsURL = source["obsURL"];
	        this.localURL = source["localURL"];
	        this.transcode = source["transcode"];
	        this.transcodeErr = source["transcodeErr"];
	        this.transcoding = source["transcoding"];
	        this.incoming = this.convertValues(source["incoming"], PathView);
	        this.outgoing = this.convertValues(source["outgoing"], PathView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snapshot {
	    settings: config.Settings;
	    environment: EnvView;
	    radio: RadioView;
	    scan: ScanView;
	    devices: DeviceView[];
	    server: ServerView;
	    session: SessionView;
	    stream: StreamView;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], config.Settings);
	        this.environment = this.convertValues(source["environment"], EnvView);
	        this.radio = this.convertValues(source["radio"], RadioView);
	        this.scan = this.convertValues(source["scan"], ScanView);
	        this.devices = this.convertValues(source["devices"], DeviceView);
	        this.server = this.convertValues(source["server"], ServerView);
	        this.session = this.convertValues(source["session"], SessionView);
	        this.stream = this.convertValues(source["stream"], StreamView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

