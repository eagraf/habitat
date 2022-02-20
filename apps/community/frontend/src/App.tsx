import {Navigation, NavItemProps} from 'react-minimal-side-navigation';
import 'react-minimal-side-navigation/lib/ReactMinimalSideNavigation.css';

import './App.css';
import Community from './AddCommunity'
import ConnectCommunity from './ConnectCommunity'
import React from 'react';

import { ConnectedCommunities, ListCommunitiesResponse } from './community';
import { AsyncState } from './types';
import axios from 'axios';
import AddMember from './AddMember';


function nameToNav(name: string): NavItemProps {
  return {title: name, itemId: '/' + name}
}

function App() {
  const [comm, setComm] = React.useState<string>("");
  const [communities, setCommunities] = React.useState<AsyncState<ListCommunitiesResponse>>({ state: 'init' });
  const [subNav, setSubNav] = React.useState<NavItemProps[]>([{title: 'Fetching communities', itemId: '', }])
  var emptyComms: Map<string, string[]> = new Map();
  const [connectedCommunities, setConnectedCommunities] = React.useState<ConnectedCommunities>(emptyComms)
    
  function GetCommunity(props: {commId: string}) {

    const blank = ["", "add", "communities"]
    if (blank.includes(props.commId)) {
      return <Community commId={props.commId} communities={connectedCommunities!} setCommunities={setConnectedCommunities}/>
    } else {
      console.log("get community for ", props.commId)
      if (props.commId && connectedCommunities?.get(props.commId)) {
        // not null or undefined: we are connected
        return <div>
          <p>Connected to community {props.commId} with addresses</p>
          <ul className="addresses">
                    {connectedCommunities.get(props.commId)!.map((addr) => (
                        <li key={addr}>
                            <h6>{addr}</h6>
                        </li>
                    ))}
                    </ul>
        </div>
      } else {
        // we have created or joined this community but not connected
        return <div>
          <ConnectCommunity commId={props!.commId} communities={connectedCommunities} setCommunities={setConnectedCommunities}></ConnectCommunity>
          <AddMember commId={props!.commId} ></AddMember>
          </div>
      }
    }
  }

  const fetchCommunities = () => {
    axios.get<ListCommunitiesResponse>(`http://localhost:8008/communities`)
          .then(response => {
              setCommunities({
                  state: 'success', 
                  data: response.data
              })
              setSubNav(response.data.Communities.map((name) => nameToNav(name)))
          }).catch((error: Error) => {
              setCommunities({
                  state: 'error',
                  message: error.message,
              })
              console.log("ERRO!!", error)
              setSubNav([{title: 'Error fetching communities!!', itemId: '', }])
          });
  };

  return (
    <>
      <Navigation
        // you can use your own router's api to get pathname
        activeItemId="/addcommunity"
        onSelect={({itemId}) => {
          // maybe push to the route
          if (itemId === '/communities') {
            fetchCommunities();
          } else {
            console.log("on select ", itemId)
            setComm(itemId)
          }
        }}
        items={[
          {
            title: 'Add Community',
            itemId: '/add',
            // you can use your own custom Icon component as well
            // icon is optional
            // elemBefore: () => <Icon name="inbox" />,
          },
          {
            title: 'Communities',
            itemId: '/communities',
            // make subNav map of communities
            subNav: subNav,
          },
        ]} />
    <GetCommunity commId={comm.substring(1)}/>
    </> 
  );
}

export default App;
