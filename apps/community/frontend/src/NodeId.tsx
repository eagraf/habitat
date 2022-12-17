import React from "react";
import axios from "axios";

type NodeIdResponse = {
  id: string;
};

const NodeId = () => {
  const [id, setId] = React.useState<string>("loading ...");
  React.useEffect(() => {
    axios
      .get<NodeIdResponse>(`http://localhost:8008/node`)
      .then((response) => {
        console.log("response getting node id", response);
        setId(response.data.id);
      })
      .catch((error: Error) => {
        console.log("error getting node id", error);
        setId(error.message);
      });
  }, []);

  return <h3> Node id is: {id} </h3>;
};

export default NodeId;
