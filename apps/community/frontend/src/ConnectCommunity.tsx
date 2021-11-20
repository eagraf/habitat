import React from 'react';
import { AsyncState } from './types'
import axios from 'axios';
import { ConnectCommunityResponse } from './community';

const ConnectCommunityContainer = () => {

    const [community, setCommunity] = React.useState<AsyncState<ConnectCommunityResponse>>({ state: 'init' });
    const [name, setName] = React.useState<string>('');

    const connectCommunity = () => {
        setCommunity({
            state: "loading",
        })
        axios.get<ConnectCommunityResponse>(`http://localhost:8008/connect?name=${name}`)
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

    const connectForm = (err: string) => {
        return (
            <div>
                <h5> {err} </h5>
                <form>
                    <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="community name" name="name" />
                    <button type="button" onClick={connectCommunity}>Connect</button>
                </form>
            </div>
        )
    }

    switch(community.state) {
        case "init":
            return connectForm('')
        case "loading":
            return <div>Connecting to community ...</div>
        case "error":
            return connectForm(`Error: ${community.message}`)
        case "success":
            return (
                <div>
                    <h5>Addresses:</h5>
                    <ul className="addresses">
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

const ConnectCommunity = () => {
    return <ConnectCommunityContainer/>
}

export default ConnectCommunity;
