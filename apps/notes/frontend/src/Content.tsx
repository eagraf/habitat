import React from 'react';
import Editor from './Editor';
import { AsyncState } from './types'
import axios from 'axios';
import { useParams, useHistory } from 'react-router-dom';
import AsyncComponent from './AsyncComponent'

const EditorContainer = (props : { file: string }) => {

    const [content, setContent] = React.useState<AsyncState<string>>({ state: 'loading' });
    React.useEffect(() => {
        setContent({
            state: "loading",
        })
        axios.get('http://localhost:8000/open?file=/' + props.file)
            .then(response => {
                console.log(response)
                setContent({
                    state: 'success',
                    data: response.data
                })
            }).catch(error => {
                setContent({
                    state: 'error',
                    message: error,
                })
            });
    }, [props.file, setContent])

    switch(content.state) {
        case "loading":
            return <div>Loading...</div>
        case "error":
            return <div>Error</div>
        case "success":
            console.log(content.data)
            return <Editor content={ content.data }/>
    }
}

const Content = () => {
    const params: { file?:string } = useParams();

    if(!params.file) {
        return <div>Click a file to open</div>
    }

    return <AsyncComponent
        request={ 'http://localhost:8000/open?file=/' + params.file }
        renderData={ (data: string) => {
            return <Editor content={ data }/>
        }}
    />
}

export default Content;
