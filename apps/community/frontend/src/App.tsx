import React from 'react';
import axios, { CancelTokenSource } from 'axios';
import './App.css';

interface CreateCommunityResponse {
  name: string
  swarm_key: string
  peer_id: string
}

const defaultCommunity:CreateCommunityResponse = {name: "fake name", swarm_key: "fake key", peer_id: "fake peer"};

const CreateCommunity = () => {
  console.log("hi 1")
  const [community, setCommunity]: [CreateCommunityResponse, (comm: CreateCommunityResponse) => void] = React.useState(
    defaultCommunity
  );
  const [loading, setLoading]: [
    boolean,
    (loading: boolean) => void
  ] = React.useState<boolean>(true);
  
  const [error, setError]: [string, (error: string) => void] = React.useState(
    ''
  );

  const cancelToken = axios.CancelToken; //create cancel token
  const [cancelTokenSource, setCancelTokenSource]: [
    CancelTokenSource,
    (cancelTokenSource: CancelTokenSource) => void
  ] = React.useState(cancelToken.source());

  const handleCancelClick = () => {
    if (cancelTokenSource) {
      cancelTokenSource.cancel('User cancelled operation');
    }
  };
  
  React.useEffect(() => {
    console.log("hi hi")
    axios.get<CreateCommunityResponse>('http://localhost:8001/ls', {
      cancelToken: cancelTokenSource.token,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 10000,
    }).then((response) => {
      console.log("respnose", response);
      setCommunity(response.data);
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
  return (
    <div className="CreateCommunityResponse">
      {loading && <button onClick={handleCancelClick}>Cancel</button>}
      <h3>{community.name}</h3>
      <h3>{community.swarm_key}</h3>
      <h3>{community.peer_id}</h3>
      {error && <p className="error">{error}</p>}
    </div>
  )
}

function App() {
  return (
    <div className="App">
      Welcome to community
      <button onClick={CreateCommunity}>Create Community</button>
      <form>
      <label>
        secret key:
        <input type="text" name="key" />
      </label>
      <label>
        bootstrap address:
        <input type="text" name="addr" />
      </label>
      <input type="submit" value="Join Community" />
    </form>
    </div>
  );
}

export default App;
