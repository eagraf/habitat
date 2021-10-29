import React from 'react';
import { useParams, useHistory } from 'react-router-dom';
import axios from 'axios';
import AsyncComponent from './AsyncComponent'

interface File {
    name: string,
    is_dir: boolean,
}

const SideMenu = () => {
    const params: { file?:string } = useParams();
    const history = useHistory();

    return <AsyncComponent 
        request={ 'http://localhost:8000/ls' }
        renderData={ (files: File[]) => {
            return <div>
                { files.map(file => {
                    return <div 
                        key={ file.name }
                        onClick={ () => history.replace("/notes/" + file.name) }
                    >
                        { params.file === file.name ? <b>{ file.name }</b> : file.name }
                    </div>
                })}
            </div>
        }}
    />
}

const SideMenuContainer = () => {
    return <div className="side-menu">
        <SideMenu />
    </div>
}

export default SideMenuContainer;
