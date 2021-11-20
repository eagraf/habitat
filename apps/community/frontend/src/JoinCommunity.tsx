import React from 'react';
import { AsyncState } from './types'
import axios from 'axios';
import { JoinCommunityResponse } from './community';

const JoinCommunityContainer = () => {

    const [community, setCommunity] = React.useState<AsyncState<JoinCommunityResponse>>({ state: 'init' });
    const [key, setKey] = React.useState<string>('');
    const [addr, setAddr] = React.useState<string>('');
    const [name, setName] = React.useState<string>('');

    const joinCommunity = () => {
        setCommunity({
            state: "loading",
        })
        axios.get<JoinCommunityResponse>(`http://localhost:8008/join?key=${key}&addr=${addr}&name=${name}`)
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

    const joinForm = (err: string) => {
        return (
            <div>
                <h5> {err} </h5>
                <form>
                    <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="community name" name="name" />
                    <input type="text" value={key} onChange={(e) => setKey(e.target.value)} placeholder="secret key" name="key" />
                    <input type="text" value={addr} onChange={(e) => setAddr(e.target.value)} placeholder="bootstrap address" name="addr" />
                    <button type="button" onClick={joinCommunity}>Join</button>
                </form>
            </div>
        )
    }

    switch(community.state) {
        case "init":
            return joinForm('')
        case "loading":
            return <div>Joining community ...</div>
        case "error":
            return joinForm(`Error: ${community.message}`)
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

const JoinCommunity = () => {
    return <JoinCommunityContainer/>
}

export default JoinCommunity;
