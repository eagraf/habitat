import React from 'react';
import Editor from './Editor';

import {
    BrowserRouter as Router,
    Route,
} from 'react-router-dom';
import Layout from './Layout';

function App() {

    return <Router>
        <Route path={"temp" /*["/notes/:file", "/notes"]*/}>
            <div className="app">
            </div>
        </Route>
        <Route path={["/notes/:file", "/notes"]}>
            <Layout />
        </Route>
    </Router>
}

export default App;
