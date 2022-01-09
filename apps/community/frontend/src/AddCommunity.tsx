import './Community.css'
import JoinCommunity from './JoinCommunity';
import CreateCommunity from './CreateCommunity';
import ConnectCommunity from './ConnectCommunity'

type Props = {
  id?: string
}

function Community(props: Props) {
  if (props.id) {
    return <p>Show page for community {props.id} </p>
  }
  return (
    <div className="Community">
        Welcome to community!
        <CreateCommunity></CreateCommunity>
        <JoinCommunity></JoinCommunity>
        <ConnectCommunity></ConnectCommunity>
      </div>
  );
}

export default Community;
