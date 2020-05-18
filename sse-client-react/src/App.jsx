import React, {useState, useEffect} from 'react';
import './App.css';
import JSONbig from 'json-bigint';
// import { LineChart, XAxis, YAxis, CartesianGrid, Tooltip, Legend, Line } from 'recharts';

function App() {

  const host = 'http://192.168.1.72:8081'

  const [data, setData] = useState({event_id:'', message:'', consumer_id:''})

  useEffect(() => {
    const serverEvent = new EventSource(`${host}/event`)

    serverEvent.onmessage = (e) => {
      const eventBig = JSONbig.parse(e.data)
      const event = JSON.parse(e.data)

      event.consumer_id = eventBig.consumer_id.toString()
      event.event_id = eventBig.event_id.toString()
      updateData(event)
    }

    return () => {
      // willUnmount
      serverEvent.close()
    }
  },[])

  const updateData = (body) => {
    setData(body)
  }

  return (
    <div>
      <div>Event(s):</div>
      {/*
      <LineChart width={600} height={300} data={data}>
        <XAxis dataKey='username'/>
        <YAxis/>
        <CartesianGrid strokeDasharray='3 3'/>
        <Tooltip/>
        <Legend />
        <Line type='monotone' dataKey='score' stroke='#8884d8' activeDot={{r: 8}}/>
      </LineChart>
      */}
      <div style={{'display': 'flex', 'flexFlow': 'column'}}>
        <span>Event ID: {data.event_id}</span>
        <span>Message: {data.message}</span>
        <span>Consumer ID: {data.consumer_id}</span>
      </div>
    </div>
  );
}

export default App
