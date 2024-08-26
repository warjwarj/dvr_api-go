import React, { useState, useEffect } from 'react';
import { Button, List, Input } from 'reactstrap';

import ApiSvrConnection from '../ApiSvrConnection';
import SignalrConnection from '../SignalrConnection';
import { getCurrentTime, formatDateTimeDVRFormat } from '../Utils';
import { TabContent, TabButtons } from './Tabs';
import VidReq from './VidReq';

export default function Devices() {

    // connections to our API server, and to the Signalr hub
    const { setReceiveCallback, apiConnection } = ApiSvrConnection;
    const { hubConnection } = SignalrConnection

    // states
    const [ msgVal, setMsgVal ] = useState("")
    const [ devList, setDevList ] = useState([]);
    const [ activeTab, setActiveTab ] = useState(0);

    // tabs
    const tabData = [
        {
            title: "Custom",
            component: null
        },
        {
            title: "Video Request",
            component: <VidReq setMsgVal={setMsgVal}/>
        },
    ];

    useEffect(() => {
        // Signalr event handlers
        hubConnection.on("UpdateDeviceList", (data) => {
            setDevList(data);
            const cmdArr = msgVal.split(';');
            const devSel = document.querySelector('#device-selector');
            cmdArr[1] = devSel.options[devSel.selectedIndex].value;
            setMsgVal(cmdArr.join(';'));
        });
        hubConnection.on("UpdateDeviceConnection", (deviceid) => {
            const option = document.createElement("option");
            const devSel = document.querySelector('#device-selector');
            option.value = deviceid;
            option.innerHTML = deviceid;
            devSel.appendChild(option);
            const cmdArr = msgVal.split(';');
            cmdArr[1] = devSel.options[devSel.selectedIndex].value;
            setMsgVal(cmdArr.join(';'));
        });
        hubConnection.on("UpdateDeviceDisconnection", (deviceid) => {
            document.querySelector(`[value="${deviceid}"]`).remove();
            const cmdArr = msgVal.split(';');
            const devSel = document.querySelector('#device-selector');
            cmdArr[1] = devSel.options[devSel.selectedIndex].value;
            setMsgVal(cmdArr.join(';'));
        });
        // API server event handlers
        setReceiveCallback((event) => {
            addMessageToLog("Incoming: " + event.data)
        });
    });

    const addMessageToLog = (message) => {
        const li = document.createElement("li");
        li.appendChild(document.createTextNode(getCurrentTime() + ": " + message));
        const firstItem = document.getElementById("message-response-list").firstChild;
        if (firstItem) {
            document.getElementById("message-response-list").insertBefore(li, firstItem);
        } else {
            document.getElementById("message-response-list").appendChild(li);
        }
    }

    const handleChange = (event) => {
        // this is the default command
        const cmdArr = msgVal.split(';');
        switch (event.target.id){
            case "device-selector":
                cmdArr[1] = event.target.options[event.target.selectedIndex].value;
                break;
            default:
                console.log("unrecognised event target id in interpretInputValue")
        }
        setMsgVal(cmdArr.join(';'));
    };

    const sendToApiSvr = () => {
        const message = document.querySelector("#send-message-input").value + "\r";
        addMessageToLog("Outgoing: " + message)
        apiConnection.send(message);
    }

    return (
        <div>
            <div className="tabs_container">
                <TabButtons
                    activeTab={activeTab}
                    setActiveTab={setActiveTab}
                    tabData={tabData}
                />
                <TabContent 
                    activeTab={activeTab}
                    tabData={tabData}
                />
            </div>
            <label htmlFor="device-selector">Device: </label>
            <Input id="device-selector" type="select" className="api-message-param" onChange={handleChange} required>
                {devList.map((id) => (
                    <option key={id} value={id}>
                    {id}
                    </option>
                ))}
            </Input>
            <br />
            <Input type="text" id="send-message-input" defaultValue={msgVal}/>
            <small>^reload the page to reset the input box.</small>
            <br />
            <Button id="send-message-button" onClick={sendToApiSvr} color="primary">Send</Button>
            <br />
            <List id="message-response-list"></List>
        </div>
    );
}