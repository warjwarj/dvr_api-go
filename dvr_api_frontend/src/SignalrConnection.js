import * as signalR from "@microsoft/signalr"
import 'dotenv/config'

const HUB_URL = process.env.DVR_API_SIGNALR_DEVICES_ENDPOINT;

class SignalrConnection {
    constructor() {
        // field
        this.hubConnection = new signalR.HubConnectionBuilder()
            .withAutomaticReconnect()
            .configureLogging(signalR.LogLevel.Debug)
            .withUrl(HUB_URL, {
                skipNegotiation: true,
                transport: signalR.HttpTransportType.WebSockets
            })
            .build();
        this.hubConnection.start().catch(err => console.error(err));
    }
    // singleton pattern to restrict instances
    static getInstance() {
        if (!SignalrConnection.instance) SignalrConnection.instance = new SignalrConnection();
        return SignalrConnection.instance;
    }
}
export default SignalrConnection.getInstance();