import React, { useEffect, useMemo, useState } from 'react'

import * as Y from 'yjs';

import * as ipfsHttpClient from 'ipfs-http-client';

import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCursor from '@tiptap/extension-collaboration-cursor'
import Placeholder from '@tiptap/extension-placeholder'

import useProvider from './useProvider';

import stringToColor from 'string-to-color'

export default function Editor() {
  const yDoc = useMemo(() => {
    console.log('ydoc')
    const yDoc = new Y.Doc()
    return yDoc
  }, [])

  const docName = 'docName'

  const ipfs = useMemo(() => {
    console.log('ipfs')
    return ipfsHttpClient.create({
      host: 'localhost',
      port: 5001,
    })
  }, [])

  const provider = useProvider(docName, yDoc, ipfs)

  const editor = useEditor({
    extensions: [
      StarterKit,
      Collaboration.configure({
        document: provider.yDoc,
      }),
      CollaborationCursor.configure({
        provider: provider,
        user: {
          name: provider.peerId.substring(0, 8),
          color: stringToColor(provider.peerId)
        }
      }),
      Placeholder.configure({
        placeholder: 'Type here...'
      })
    ],
  })

  useEffect(() => {
    provider.connect()
  }, [provider])

  return <div>
    <EditorContent editor={editor} />
  </div>

}