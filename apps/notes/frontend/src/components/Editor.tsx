import React, { useEffect, useMemo, useState } from 'react'

import * as Y from 'yjs';

import * as ipfsHttpClient from 'ipfs-http-client';

import { useEditor, EditorContent, Editor as TipTapEditor } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCursor from '@tiptap/extension-collaboration-cursor'
import Placeholder from '@tiptap/extension-placeholder'

import stringToColor from 'string-to-color'
import WebRtcProvider from '../yjsProviders/WebRtcProvider';
import BackendProvider from '../yjsProviders/BackendProvider';
import useAsync from '../util/useAsync';
import { AsyncState } from '../types/asyncTypes';

import { EditorView } from 'prosemirror-view'


/**
 * Weird hack to remove bug that happens when switching between files after editing
 * from: https://github.com/ueberdosis/tiptap/issues/1451#issuecomment-953348865
 */
EditorView.prototype.updateState = function updateState(state) {
  if (!this.docView) return // This prevents the matchesNode error on hot reloads
  this.updateStateInner(state, this.state.plugins != state.plugins)
}

interface EditorContainerProps {
  docName: string
}

const ipfs = ipfsHttpClient.create({
  host: 'localhost',
  port: 5001,
})

const backend_url = 'localhost:' + (window.location.search ? window.location.search.substring(1) : 4000)
console.log(backend_url)

export default function EditorContainer({ docName }: EditorContainerProps) {

  const [docState, setDocState] = useState(() => {
    const yDoc = new Y.Doc()
    return {
      yDoc,
      webrtcProvider: new WebRtcProvider(docName, yDoc, ipfs),
      backendProvider: new BackendProvider(docName, yDoc, backend_url)
    }
  })

  useEffect(() => {
    docState.webrtcProvider.connect()
    docState.backendProvider.connect()
  }, [])

  const editor = useEditor({
    extensions: [
      StarterKit,
      Collaboration.configure({
        document: docState.webrtcProvider.yDoc,
      }),
      CollaborationCursor.configure({
        provider: docState.webrtcProvider,
        user: {
          name: docState.webrtcProvider.peerId.substring(0, 8),
          color: stringToColor(docState.webrtcProvider.peerId)
        },
      }),
      Placeholder.configure({
        placeholder: 'Type here...'
      })
    ],
  }, [docState])

  useEffect(() => {
    if (docState.webrtcProvider.docName !== docName) {
      editor.destroy()
      docState.webrtcProvider.disconnect()
      docState.backendProvider.disconnect()
      docState.yDoc.destroy()

      const yDoc = new Y.Doc()
      const webrtcProvider = new WebRtcProvider(docName, yDoc, ipfs)
      const backendProvider = new BackendProvider(docName, yDoc, backend_url)
      setDocState({
        yDoc,
        webrtcProvider,
        backendProvider
      })

      webrtcProvider.connect()
      backendProvider.connect()
    }
  }, [docName])

  return <div 
      className="notes-editor-container"
      onClick={() => {
        editor.commands.focus()
      }}
    >
    <EditorContent
      editor={editor}
      className="notes-editor"
    />
  </div>
}