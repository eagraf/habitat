import React from 'react';
import Immutable from 'immutable';
import { 
    Editor as DraftEditor, 
    EditorState,
    RichUtils,
    ContentState,
    ContentBlock,
    DefaultDraftBlockRenderMap,
    Modifier,
    EditorBlock,
} from 'draft-js';
import 'draft-js/dist/Draft.css';

const resetOnSplitBlockTypes = [
    'title',
    'header-one',
    'header-two',
    'header-three',
]

function Title(props) {
    console.log(props);
    return <h1>
        { 
            !props.block.getText().trim() ? 
            <div className="title-placeholder">Title...</div> : 
            null 
        }
        <EditorBlock
            {...props}
        >
        </EditorBlock>
    </h1>
}

function blockRenderer(contentBlock) {
    const type = contentBlock.getType();
    switch(type) {
        case 'title': {
            return { 
                component: Title 
            }
        }
    }
}

const blockRenderMap = Immutable.Map({
})

const extendedBlockRenderMap = DefaultDraftBlockRenderMap.merge(blockRenderMap);

function blockStyle(contentBlock) {
    const type = contentBlock.getType();
    switch(type) {
    }
}

function Editor() {
    const [editorState, setEditorState] = React.useState(
        () => EditorState.createWithContent(ContentState.createFromBlockArray([
            new ContentBlock({
                text: 'Hello world',
                type: 'title',
                key: 'titleKey'
            })
        ])
    ));

    function toggleStyle(type) {
        return () => {
            if(RichUtils.getCurrentBlockType(editorState) != 'title') {
                const toggleStyle = RichUtils.toggleBlockType(editorState, type)
                const refocus = EditorState.moveFocusToEnd(toggleStyle);
                setEditorState(refocus);
            }
            else {
                setEditorState(EditorState.moveFocusToEnd(editorState));
            }
        }
    }

    return <div className="editor">
        <div>
            <button onClick={ toggleStyle('header-two') }>H1</button>
            <button onClick={ toggleStyle('header-three') }>H2</button>
            <button onClick={ toggleStyle('header-four') }>H3</button>
            <button onClick={ toggleStyle('ordered-list-item') }>List</button>
            <button onClick={ toggleStyle('code-block') }>Code</button>
        </div>
        <DraftEditor 
            editorState={ editorState } 
            onChange={ newState => {
                if(newState.getCurrentContent().getBlockForKey('titleKey').getType() !== 'title') {
                    //TODO: update newState with title
                } 
                setEditorState(newState) 
            }}
            handleKeyCommand={ (command, editorState) => {
                console.log(command);
                const currentContent = editorState.getCurrentContent(); 
                const currentSelection = editorState.getSelection();
                const currentKey = currentSelection.getStartKey();
                const currentBlock = currentContent.getBlockForKey(currentKey);
                if(command === 'split-block') {
                    if(resetOnSplitBlockTypes.includes(currentBlock.getType())) {
                        const splitBlock = EditorState.push(
                            editorState,
                            Modifier.splitBlock(
                                currentContent,
                                currentSelection,
                            )
                        )
                        const changeStyle = EditorState.push(
                            splitBlock,
                            Modifier.setBlockType(
                                splitBlock.getCurrentContent(), 
                                splitBlock.getSelection(),
                                'unstyled'
                            ),
                            'change-block-type'
                        )
                        setEditorState(changeStyle)
                        return 'handled';
                    }
                }
                const newState = RichUtils.handleKeyCommand(editorState, command);
                if(newState) {
                    if(newState.getCurrentContent().getBlockForKey('titleKey').getType() !== 'title') {
                        return 'handled';
                    }
                    setEditorState(newState);
                    return 'handled';
                }

                return 'not-handled';
            }}
            blockRendererFn={ blockRenderer }
            blockStyleFn={ blockStyle }
            blockRenderMap={ extendedBlockRenderMap }
        />
    </div>
};

export default Editor;
