/*
 * Copyright (c) Joseph Prichard 2023
 */

import './App.css';
import { Router, Route, Routes } from '@solidjs/router';
import Home from './pages/Home';
import Room from './pages/Room';
import NotificationPanel, { defaultTempMsg, TempMsg, useTempMsg } from "./components/NotificationPanel";
import { Context, createContext } from "solid-js";

export const WS_URL = "ws://localhost:8080";
export const BACKEND_URL = "http://localhost:8080/api";
export const TempMsgContext: Context<TempMsg> = createContext(defaultTempMsg());

const App = () => {
    const tempMsg = useTempMsg(5000);

    return (
        <div class="App">
            <TempMsgContext.Provider value={tempMsg}>
                <Router>
                    <NotificationPanel errorMsg={tempMsg.msg} onClose={tempMsg.clearMsg}/>
                    <Routes>
                        <Route path="/" component={Home}/>
                        <Route path="/rooms/:code" component={Room}/>
                    </Routes>
                </Router>
            </TempMsgContext.Provider>
        </div>
    );
};

export default App;
