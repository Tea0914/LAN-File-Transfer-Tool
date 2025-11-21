export namespace main {
	
	export class TransferStats {
	    totalFiles: number;
	    completedFiles: number;
	    totalBytes: number;
	    transferredBytes: number;
	    currentSpeed: number;
	    estimatedTime: string;
	    currentFile: string;
	    progress: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalFiles = source["totalFiles"];
	        this.completedFiles = source["completedFiles"];
	        this.totalBytes = source["totalBytes"];
	        this.transferredBytes = source["transferredBytes"];
	        this.currentSpeed = source["currentSpeed"];
	        this.estimatedTime = source["estimatedTime"];
	        this.currentFile = source["currentFile"];
	        this.progress = source["progress"];
	        this.status = source["status"];
	    }
	}

}

