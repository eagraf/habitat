import "./Community.css";
import JoinCommunity from "./JoinCommunity";
import CreateCommunity from "./CreateCommunity";
import { ConnectedCommunities } from "./community";

type Props = {
  commId: string;
  communities: ConnectedCommunities;
  setCommunities: React.Dispatch<React.SetStateAction<ConnectedCommunities>>;
};

function Community(props: Props) {
  return (
    <div
      className="Community"
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <p className="CommunityInput" style={{ paddingTop: "1%" }}>
        Welcome to community!
      </p>
      <CreateCommunity></CreateCommunity>
      <JoinCommunity></JoinCommunity>
    </div>
  );
}

export default Community;
