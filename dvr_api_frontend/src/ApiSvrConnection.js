import 'dotenv/config'

const URL = process.env.DVR_API_API_ENDPOINT;

class ApiSvrConnection {
    constructor(){
        this.connect();
        this.timeout = 1000;
    }

    setReceiveCallback = (callback) => {
        this.processReceived = callback
    }
    
    connect = () => {
        this.apiConnection = new WebSocket(URL)
        this.apiConnection.onopen = (event) => {
            console.log("WS connected to API server");
            this.apiConnection.send("Hi server, I'm an API client!");
        }
        this.apiConnection.onmessage = (event) => {
            this.processReceived(event);
        }
        this.apiConnection.onclose = (event) => {
            console.log("WS disconnected from API server");
            delete this.apiConnection; // gc
            setTimeout(this.connect, this.timeout += this.timeout)
        }
    }
    // singleton pattern to restrict instances
    static getInstance() {
        if (!ApiSvrConnection.instance) ApiSvrConnection.instance = new ApiSvrConnection();
        return ApiSvrConnection.instance;
    }
}
export default ApiSvrConnection.getInstance();

// this.apiSvrConnection = new WebSocket(URL);
// this.apiSvrConnection.onopen = (event) => {
//     console.log("Connected to API server: ", event.data)
// }
// this.apiSvrConnection.onerror = (event) => {
//     console.log("Error in websocket connection to API server: ", event.data)
// }
// this.apiSvrConnection.onclose = (event) => {
//     console.log("Disconnected from API server: ", event.data);
// }