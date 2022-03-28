import React from "react";
import { AsyncState } from "./types";
import axios from "axios";
import { ConnectCommunityResponse } from "./community";
import "./Community.css";

const CreateCommunityContainer = () => {
  const [community, setCommunity] = React.useState<
    AsyncState<ConnectCommunityResponse>
  >({ state: "init" });
  const [name, setName] = React.useState<string>("");
  const [id, setId] = React.useState<string>("");
  const [createIpfs, setCreateIpfs] = React.useState<boolean>(false);

  const createCommunity = () => {
    setCommunity({
      state: "loading",
    });
    axios
      .get<ConnectCommunityResponse>(
        `http://localhost:8008/create?name=${name}&id=${id}&ipfs=${createIpfs}`
      )
      .then((response) => {
        console.log("respppp ", response);
        setCommunity({
          state: "success",
          data: response.data,
        });
        console.log("got response", response.data);
      })
      .catch((error: Error) => {
        console.log("errrorrr ", error);
        setCommunity({
          state: "error",
          message: error.message,
        });
        console.error("got error", error.message);
      });
  };

  const createForm = (err: string) => {
    return (
      <div>
        <h5> {err} </h5>
        <form>
          <div>
            <input
              type="radio"
              value="true"
              name="create_ipfs"
              onChange={(_) => setCreateIpfs(true)}
            />{" "}
            First Time
            <input
              type="radio"
              value="false"
              name="raft_only"
              onChange={(_) => setCreateIpfs(false)}
            />{" "}
            Join Existing (raft only)
          </div>
          <input
            type="text"
            name="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="community name"
          />
          <input
            type="text"
            name="id"
            value={id}
            onChange={(e) => setId(e.target.value)}
            placeholder="(optional) community id"
          />
          <button type="button" onClick={createCommunity}>
            Create
          </button>
        </form>
      </div>
    );
  };

  switch (community.state) {
    case "init":
      return createForm("");
    case "loading":
      return <div>Joining community ...</div>;
    case "error":
      return createForm(`Error: ${community.message}`);
    case "success":
      return (
        <div className="CommunityInput">
          <h5>name: {community.data.Name}</h5>
          <h5>id: {community.data.CommId}</h5>
          <h5>PeerId: {community.data.PeerId}</h5>
          <h5>swarm key: {community.data.SwarmKey}</h5>
          <h5>addrs:</h5>
          <ul className="addrs">
            {community.data.Addresses.map((addr) => (
              <li key={addr}>
                <h6>{addr}</h6>
              </li>
            ))}
          </ul>
        </div>
      );
  }
};

const CreateCommunity = () => {
  return <CreateCommunityContainer />;
};

export default CreateCommunity;
