export namespace types {
	
	export class DeviceInfo {
	    UUID: string;
	    PublicKey: string;
	    EK: number[];
	    AIK: number[];
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UUID = source["UUID"];
	        this.PublicKey = source["PublicKey"];
	        this.EK = source["EK"];
	        this.AIK = source["AIK"];
	    }
	}
	export class TPMStatus {
	    available: boolean;
	    initialized: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TPMStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.initialized = source["initialized"];
	    }
	}

}

