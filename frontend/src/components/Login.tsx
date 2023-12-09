/*
 * Copyright (c) Joseph Prichard 2023
 */

import "./Login.css";
import { gapi } from 'gapi-script';
import { onMount } from "solid-js";

const Login = () => {

    console.log(gapi)

    onMount(() => {
        gapi.signin2.render("sign-in2", {
            'scope': 'profile email',
            'width': 240,
            'height': 50,
            'longtitle': true,
            'theme': 'dark',
            'onsuccess': () => {},
            'onfailure': () => {},
        });
    });

    return (
        // <div id="sign-in2" class="g-signin2 Login"/>
        <div></div>
    );
}

export default Login;