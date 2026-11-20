export namespace config {
	
	export class Settings {
	    rtmpPort: number;
	    apiPort: number;
	    mediaMtxPath: string;
	    ffmpegPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rtmpPort = source["rtmpPort"];
	        this.apiPort = source["apiPort"];
	        this.mediaMtxPath = source["mediaMtxPath"];
	        this.ffmpegPath = source["ffmpegPath"];
	    }
	}

}

export namespace main {
	
	export class EnvView {
	    os: string;
	    arch: string;
	    localIP: string;
	    dataDir: string;
	    mediaMtxPath: string;
	    ffmpegPath: string;
	    mediaMtxReady: boolean;
	    ffmpegReady: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EnvView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.localIP = source["localIP"];
	        this.dataDir = source["dataDir"];
	        this.mediaMtxPath = source["mediaMtxPath"];
	        this.ffmpegPath = source["ffmpegPath"];
	        this.mediaMtxReady = source["mediaMtxReady"];
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
	export class StreamView {
	    host: string;
	    rtmpPort: number;
	    directURL: string;
	    webrtcPort: number;
	    webrtcURL: string;
	    webrtcLocalURL: string;
	    webrtcReady: boolean;
	    webrtcAudio: boolean;
	    encoderErr: string;
	    incoming: PathView;
	    webrtc: PathView;
	
	    static createFrom(source: any = {}) {
	        return new StreamView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.rtmpPort = source["rtmpPort"];
	        this.directURL = source["directURL"];
	        this.webrtcPort = source["webrtcPort"];
	        this.webrtcURL = source["webrtcURL"];
	        this.webrtcLocalURL = source["webrtcLocalURL"];
	        this.webrtcReady = source["webrtcReady"];
	        this.webrtcAudio = source["webrtcAudio"];
	        this.encoderErr = source["encoderErr"];
	        this.incoming = this.convertValues(source["incoming"], PathView);
	        this.webrtc = this.convertValues(source["webrtc"], PathView);
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
	    server: ServerView;
	    stream: StreamView;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], config.Settings);
	        this.environment = this.convertValues(source["environment"], EnvView);
	        this.server = this.convertValues(source["server"], ServerView);
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

