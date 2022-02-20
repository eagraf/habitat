import React from 'react';
import { AsyncState } from './types'
import axios from 'axios';
import { ConnectCommunityResponse, JoinCommunityResponse } from './community';

const JoinCommunityContainer = () => {

    const [community, setCommunity] = React.useState<AsyncState<ConnectCommunityResponse>>({ state: 'init' });
    const [key, setKey] = React.useState<string>('');
    const [addr, setAddr] = React.useState<string>('');
    const [name, setName] = React.useState<string>('');
    const [comm, setComm] = React.useState<string>('');

    const joinCommunity = () => {
        setCommunity({
            state: "loading",
        })
        axios.get<ConnectCommunityResponse>(`http://localhost:8008/join?key=${key}&addr=${addr}&name=${name}&comm=${comm}`)
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
            <div className='CommunityInput'>
                <h5> {err} </h5>
                <form>
                    <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="community name" name="name" />
                    <input type="text" value={key} onChange={(e) => setKey(e.target.value)} placeholder="secret key" name="key" />
                    <input type="text" value={addr} onChange={(e) => setAddr(e.target.value)} placeholder="bootstrap address" name="addr" />
                    <input type="text" value={comm} onChange={(e) => setComm(e.target.value)} placeholder="community id" name="comm" />
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
                <div className='CommunityInput'>
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
            )
    }
}

const JoinCommunity = () => {
    return <JoinCommunityContainer/>
}

export default JoinCommunity;
