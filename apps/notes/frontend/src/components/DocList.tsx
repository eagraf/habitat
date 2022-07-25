import React from "react";
import useAsync from "../util/useAsync";

const backend_url = 'http://localhost:' + (window.location.search ? window.location.search.substring(1) : 4000)


export default function DocList({ notesState, dispatch }) {
    const asyncState = useAsync<string[]>(async () => {
        const response = await fetch(backend_url + '/docs')

        const docList = await response.json()

        dispatch({ type: 'SET_DOCS', docList })

        return null
    }, [dispatch])
    console.log(notesState)
    switch (asyncState.state) {
        case 'loading':
            return <div className="loading"></div>
        case 'error':
            return <div>Unable to load documents</div>
        case 'success':
            return <>
                <h2>NOTES: </h2>
                <div className="notes-doclist">
                    <ul className="list-group list-group-flush">
                        { notesState.docList.map(d => {
                            const className = "notes-doclist-item list-group-item" + (d === notesState.currentDoc ? " notes-doclist-item-selected" : "")
                            return <a href="#" key={d} className={className} onClick={ () => dispatch({ type: 'UPDATE_CURRENT_DOC', doc: d })}>
                                <li >
                                    {d.split("/")[3]}
                                </li>
                            </a>
                        })}
                    </ul>
                </div>
            </>
    }
}