import React from 'react';
import axios, { CancelTokenSource } from 'axios';
import './App.css';
import SideMenu from './SideMenu';
import Editor from './Editor';

interface ListFilesResponse {
  name: string
  is_dir: boolean
  filemode: number
}

const defaultFiles:ListFilesResponse[] = [];

const FileList = () => {
  const [files, setFiles]: [ListFilesResponse[], (files: ListFilesResponse[]) => void] = React.useState(
    defaultFiles
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
    console.log("HELLO");
    axios
      .get<ListFilesResponse[]>('http://localhost:8000/ls', {
        cancelToken: cancelTokenSource.token,
        headers: {
          'Content-Type': 'application/json',
        },
        timeout: 10000,
      })
      .then((response) => {
        console.log("WAUT");
        console.log(response.data);
        setFiles(response.data);
        setLoading(false);
      })
      .catch((ex) => {
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
  }, []);

  return (
    <div className="FilesList">
      {loading && <button onClick={handleCancelClick}>Cancel</button>}
      <ul className="files">
        {files.map((file) => (
          <li key={file.name}>
            <h3>{file.name}</h3>
          </li>
        ))}
      </ul>
      {error && <p className="error">{error}</p>}
    </div>
  );
}

function App() {
  return (
    <div className="app">
        <SideMenu />
        <Editor />
    </div>
  );
}

export default App;
