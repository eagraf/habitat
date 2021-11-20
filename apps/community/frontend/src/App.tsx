import './App.css';
import JoinCommunity from './JoinCommunity';
import CreateCommunity from './CreateCommunity';
import ConnectCommunity from './ConnectCommunity'

function App() {
  return (
    <div className="App">
      Welcome to community!
      <CreateCommunity></CreateCommunity>
      <JoinCommunity></JoinCommunity>
      <ConnectCommunity></ConnectCommunity>
    </div>
  );
}

export default App;
