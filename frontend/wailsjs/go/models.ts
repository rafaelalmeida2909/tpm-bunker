export namespace types {
	
	export class APIResponse {
	    Success: boolean;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new APIResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Message = source["Message"];
	    }
	}
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
	export class UserOperation {
	    Type: string;
	    Data: any;
	
	    static createFrom(source: any = {}) {
	        return new UserOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Data = source["Data"];
	    }
	}

}

