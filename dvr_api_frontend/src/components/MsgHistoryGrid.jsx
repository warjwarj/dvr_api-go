import { AgGridReact } from 'ag-grid-react';
import "ag-grid-community/styles/ag-grid.css";

import { useState, useEffect } from 'react'
import { fetchMsgHistory } from '../HttpApiConn';

const testRequest = {
  "after": "2023-10-03T16:45:14.000+00:00",
  "before": "2025-10-03T16:45:14.000+00:00",
  "devices": [
      "123456",
  ]
}

export function MsgHistoryGrid() {

    // Row Data: The data to be displayed. TODO - multiple devices
    const [rowData, setRowData] = useState('');

    useEffect(() => {
      setRowData(fetchMsgHistory(testRequest).then((data) => {
        console.log(data)
      }))
    }, [])

    // how we organise the data
    const columns = {
      field: "direction",
      field: "message",
      field: "packet_time",
      field: "received_time",
    }
    
   
    return (
        // wrapping container with theme & size
        <div
         className="ag-theme-quartz" // applying the Data Grid theme
         style={{ height: 500 }} // the Data Grid will fill the size of the parent container
        >
          <AgGridReact
              rowData={rowData[0]["MsgHistory"]}
              columnDefs={columns}
          />
        </div>
    )
}