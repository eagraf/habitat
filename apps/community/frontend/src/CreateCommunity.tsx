import React from 'react';
import { AsyncState } from './types'
import axios from 'axios';
import { CreateCommunityResponse } from './community';

const CreateCommunityContainer = () => {

    const [community, setCommunity] = React.useState<AsyncState<CreateCommunityResponse>>({ state: 'init' });
    const [name, setName] = React.useState<string>('');

    const createCommunity = () => {
        setCommunity({
            state: "loading",
        })
        axios.get<CreateCommunityResponse>(`http://localhost:8008/create?name=${name}`)
            .then(response => {
                setCommunity({
                    state: 'success',
                    data: response.data
                })
            }).catch((error: Error) => {
                setCommunity({
                    state: 'error',
                    message: error.message,
                })
            });
    };

    const createForm = (err: string) => {
        return (
            <div>
                <h5> {err} </h5>
                <form>
                    <input type="text" name="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="community name" />
                    <button type="button" onClick={createCommunity}>Create</button>
                </form>
            </div>
            
        )
        
    }

    switch(community.state) {
        case "init":
            return createForm('')
        case "loading":
            return <div>Joining community ...</div>
        case "error":
            return createForm(`Error: ${community.message}`)
        case "success":
            return (
                <div>
                    <h5>name: {community.data.name}</h5>
                    <h5>swarm key: {community.data.swarm_key}</h5>
                    <h5>bootstrap peers:</h5>
                    <ul className="btstp_peers">
                    {community.data.btstp_peers.map((peer) => (
                        <li key={peer}>
                            <h6>{peer}</h6>
                        </li>
                    ))}
                    </ul>
                    <h5>peers:</h5>
                    <ul className="peers">
                    {community.data.peers.map((peer) => (
                        <li key={peer}>
                            <h6>{peer}</h6>
                        </li>
                    ))}
                    </ul>
                </div>
            )
    }
}

const CreateCommunity = () => {
    return <CreateCommunityContainer/>
}

export default CreateCommunity;
