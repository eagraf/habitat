import React from 'react';
import './App.css';
import SideMenu from './SideMenu';
import Editor from './Editor';

import { 
    BrowserRouter as Router,
    Route,
} from 'react-router-dom';

function App() {
    return <Router>
        <Route path={ "temp" /*["/notes/:file", "/notes"]*/ }>
            <div className="app">
            </div>
        </Route>
        <Route path={ ["/notes/:file", "/notes"] }>
            <div className="app">
                <Editor />
            </div>
        </Route>
    </Router>
}

export default App;
