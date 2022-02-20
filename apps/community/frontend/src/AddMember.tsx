import React from 'react';
import { AsyncState } from './types'
import axios from 'axios';
import { AddMemberResponse } from './community';
import './Community.css'

type Props = {
    commId: string
}

const AddMemberContainer = (props: Props) => {

    const [memberData, setMemberData] = React.useState<AsyncState<AddMemberResponse>>({ state: 'init' });
    const [member, setMember] = React.useState<string>('');

    const addMember = () => {
        setMemberData({
            state: "loading",
        })
        axios.get<AddMemberResponse>(`http://localhost:8008/add?comm=${props.commId}&node=${member}`)
            .then(response => {
                console.log("respppp ", response)
                setMemberData({
                    state: 'success',
                    data: response.data
                })
                console.log("got response", response.data)
            }).catch((error: Error) => {
                console.log("errrorrr ", error)
                setMemberData({
                    state: 'error',
                    message: error.message,
                })
                console.error("got error", error.message)
            });
    };

    const createForm = (err: string) => {
        return (
            <div>
                <h5> {err} </h5>
                <form>
                    <input type="text" name="member" value={member} onChange={(e) => setMember(e.target.value)} placeholder="node id" />
                    <button type="button" onClick={addMember}>Add Member</button>
                </form>
            </div>
            
        )
        
    }

    switch(memberData.state) {
        case "init":
            return createForm('')
        case "loading":
            return <div>Adding node to community ...</div>
        case "error":
            return createForm(`Error: ${memberData.message}`)
        case "success":
            return (
                <div className='Member Data'>
                    <h5>added node to community: {memberData.data.MemberId}</h5>
                    <h5>node id: {memberData.data.NodeId}</h5>
                </div>
            )
    }
}

function AddMember(props: Props) {
    return <AddMemberContainer commId={props.commId}/>
}

export default AddMember;
