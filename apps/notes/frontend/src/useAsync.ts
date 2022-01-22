import { useState, useEffect } from 'react';
import { AsyncState } from './types'

export default function useAsync<T>(async: () => Promise<T>, dependencies?: any[], cleanup?: () => void): AsyncState<T> {
    const [state, setState] = useState<AsyncState<T>>({ state: 'loading' });

    useEffect(() => {
        async().then(result => {
            setState({ 
                state: 'success',
                data: result
            })
        }).catch(error => {
            console.error(error)
            setState({
                state: 'error',
                message: error.message || error,
            })
        })
        return cleanup
    }, dependencies || [])

    return state;
}