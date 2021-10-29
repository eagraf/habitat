import React from 'react';
import axios from 'axios';
import { AsyncState } from './types'

interface AsyncComponentProps<T> {
    request: string,
    handleError?: (error: string) => string,
    handleResponse?: (response: object) => T,
    renderError?: (error: string) => React.ReactElement,
    renderLoading?: () => React.ReactElement,
    renderData: (data: T) => React.ReactElement,
}

function AsyncComponent<T>(props: AsyncComponentProps<T>) {
    const [state, setState] = React.useState<AsyncState<T>>({ state: 'loading' });
    
    React.useEffect(() => {
        setState({
            state: "loading",
        })
        axios.get(props.request)
            .then(response => {
                console.log(response)
                if(props.handleResponse) {
                    setState({
                        state: 'success',
                        data: props.handleResponse(response.data)
                    })
                }
                else {
                    setState({
                        state: 'success',
                        data: response.data
                    })
                }
            }).catch(error => {
                if(props.handleError) {
                    setState({
                        state: 'error',
                        message: props.handleError(error),
                    })
                }
                else {
                    setState({
                        state: 'error',
                        message: error,
                    })
                }
            });
    }, [
        props.request, 
        props.handleResponse, 
        props.handleError, 
        setState
    ]);

    switch(state.state) {
        case "loading":
            if(props.renderLoading) {
                return props.renderLoading();
            }
            else {
                return <div>Loading...</div>;
            }
        case "error":
            if(props.renderError) {
                return props.renderError(state.message);
            }
            else {
                return <div>Error: { state.message }</div>
            }
        case "success":
            return props.renderData(state.data)
    }
}

export default AsyncComponent;
