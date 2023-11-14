import './App.css';
import { Router, Route, Routes } from '@solidjs/router';
import Home from './pages/Home';
import Room from './pages/Room';

export const DOMAIN = "localhost:8080";

const App = () => {
    return (
        <div class="App">
            <Router>
                <Routes>
                    <Route path="/" component={Home}/>
                    <Route path="/rooms/:code" component={Room}/>
                </Routes>
            </Router>
        </div>
    );
};

export default App;
