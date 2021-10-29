import React from 'react';
import './App.css';
import SideMenu from './SideMenu';
import Editor from './Editor';
import Content from './Content';
import { 
    BrowserRouter as Router,
    Route,
} from 'react-router-dom';

function App() {
    return <Router>
        <Route path={ ["/notes/:file", "/notes"] }>
            <div className="app">
                <SideMenu />
                <Content />
            </div>
        </Route>
    </Router>
}

export default App;
