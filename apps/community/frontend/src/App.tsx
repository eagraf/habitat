import {Navigation} from 'react-minimal-side-navigation';
import 'react-minimal-side-navigation/lib/ReactMinimalSideNavigation.css';

import './App.css';
import Community from './AddCommunity'
import React from 'react';

function GetCommunity(props: {commId: string}) {
  console.log("get community for ", props.commId)
  const blank = ["/add", "/communities"]
  if (blank.includes(props.commId)) {
    return <Community/>
  } else {
    return <Community id={props.commId}/>
  }
}

function App() {
  const [comm, setComm] = React.useState<string>("");
  return (
    <>
      <Navigation
        // you can use your own router's api to get pathname
        activeItemId="/addcommunity"
        onSelect={({itemId}) => {
          // maybe push to the route
          setComm(itemId)
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
            subNav: [
              {
                title: 'Put existing communities here',
                itemId: '/somecommunity',
              },
            ],
          },
        ]} />
    <GetCommunity commId={comm}/>
    </> 
  );
}

export default App;
