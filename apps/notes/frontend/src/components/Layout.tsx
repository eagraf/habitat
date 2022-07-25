import React from "react"
import DocList from "./DocList"
import Editor from "./Editor"
import { Form, Button, Modal } from 'react-bootstrap'

const backend_url = 'http://localhost:' + (window.location.search ? window.location.search.substring(1) : 4000)

const initialNotesState = {
    currentDoc: '/orbitdb/zdpuAwpxc5d1CZ1VgxSDW7sFzNrRBe3ZwXiRuHjJ1gM7FiZbA/docName',
    docList: []
}

function notesReducer(state, action) {
    switch (action.type) {
        case 'UPDATE_CURRENT_DOC':
            return {
                ...state,
                currentDoc: action.doc
            }
        case 'ADD_DOC':
            return {
                ...state,
                docList: state.docList.concat([action.doc])
            }
        case 'SET_DOCS':
            return {
                ...state,
                docList: action.docList
            }
        default:
            console.log(action.type)
            throw new Error()
    }
}

export default function Layout() {
    const [modal, setModal] = React.useState(null)
    const [state, dispatch] = React.useReducer(notesReducer, initialNotesState)

    return <div className="notes-app">
        <div className="notes-menu">
            <h1>🌱</h1>
            <DocList notesState={state} dispatch={dispatch} />
            <h2>Actions:</h2>
            <div className="notes-new-buttons-container">
                <Button onClick={() => setModal('new')} className="notes-icon-button">
                    <i className="bi-plus-lg"></i>
                    <span>New Note</span>
                </Button>
                <Button onClick={() => setModal('join')} className="notes-icon-button">
                    <i className="bi-people"></i>
                    <span>Join Note</span>
                </Button>
            </div>
        </div>
        <Editor docName={state.currentDoc} />

        <Modal show={modal === 'new'} onHide={() => setModal(null)} restoreFocus={false} centered>
            <Modal.Header closeButton>
                <Modal.Title>New Note</Modal.Title>
            </Modal.Header>
            <Form onSubmit={async (event) => {
                event.preventDefault()
                //@ts-ignore
                const noteName = event.target.noteName.value
                const response = await fetch(backend_url + '/newDoc?name=' + noteName, { method: 'POST' })
                const addr = await response.text()
                console.log(addr)
                dispatch({ type: 'ADD_DOC', doc: addr })
                dispatch({ type: 'UPDATE_CURRENT_DOC', doc: addr })

                setModal(null)
            }}>

                <Modal.Body>
                    <Form.Group>
                        <Form.Control name="noteName" placeholder="Enter note name" />
                    </Form.Group>
                </Modal.Body>

                <Modal.Footer>
                    <Button type="submit" className="notes-button">Create</Button>
                </Modal.Footer>
            </Form>

        </Modal>

        <Modal show={modal === 'join'} onHide={() => setModal(null)} restoreFocus={false} centered>
            <Modal.Header closeButton>
                <Modal.Title>Join Note</Modal.Title>
            </Modal.Header>
            <Form onSubmit={event => {
                event.preventDefault()
                //@ts-ignore
                const addr = event.target.noteAddr.value
                dispatch({ type: 'ADD_DOC', doc: addr })
                dispatch({ type: 'UPDATE_CURRENT_DOC', doc: addr })

                setModal(null)

            }}>

                <Modal.Body>
                    <Form.Group style={{marginBottom: 24 }}> 
                        <Form.Label>Current note address</Form.Label>
                        <Form.Text style={{ overflowWrap: 'break-word' }}>{state.currentDoc}</Form.Text>
                    </Form.Group>

                    <Form.Group>
                        <Form.Label>Join note</Form.Label>
                        <Form.Control name="noteAddr" placeholder="Enter note address" />
                    </Form.Group>
                </Modal.Body>

                <Modal.Footer>
                    <Button type="submit" className="notes-button">Join</Button>
                </Modal.Footer>
            </Form>

        </Modal>

    </div>
}