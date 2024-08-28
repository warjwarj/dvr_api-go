

const API_SVR_ENDPOINT = "ws://127.0.0.1:9046"

class ApiSvrConnection {
    constructor(){
        this.connect();
        this.timeout = 1000;
    }

    // set the function which is called when data is received
    setReceiveCallback = (callback) => {
        this.processReceived = callback
    }
    
    // called when imported, in the constructor
    connect = () => {
        this.apiConnection = new WebSocket(API_SVR_ENDPOINT, ["dvr_api"])
        this.apiConnection.onopen = (event) => {
            console.log("WS connected to API server");
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