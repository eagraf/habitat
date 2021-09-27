import React from 'react';
import axios from 'axios';
import './App.css';

interface CreateCommunityResponse {
  name: string
  swarm_key: string
  peer_id: string
}

const [loading, setLoading]: [
  boolean,
  (loading: boolean) => void
] = React.useState<boolean>(true);

const [error, setError]: [string, (error: string) => void] = React.useState(
  ''
);



const handleCreateClick = () => {
  React.useEffect(() => {
    axios.get<CreateCommunityResponse>('http://localhost:8001/ls', {
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 10000,
    }).then((response) => {
      showCommunity(response.data);
      setLoading(false);
    }).catch((ex) => {
      console.log(ex);
      if (ex.response) {
        let error = axios.isCancel(ex)
        ? 'Request Cancelled'
        : ex.code === 'ECONNABORTED'
        ? 'A timeout has occurred'
        : ex.response.status === 404
        ? 'Resource Not Found'
        : 'An unexpected error has occurred';
  
        setError(error);
      }
      setLoading(false);
      console.log(ex.request);
      setError("failed to load files");
      });
  })
  
}

function App() {
  return (
    <div className="App">
      <button onClick={}
    </div>
  );
}

export default App;
