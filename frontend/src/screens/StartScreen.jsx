import {connect, sendMsg} from "../api/index.js";
import {useEffect, useState} from "react";

export default function StartScreen({ setScreen, message }) {
    
    function RollDice(){
        const payload = JSON.stringify({
            action: "ROLL_DICE"
        });
        sendMsg(payload);
    }

    return (
        <>
            <h1>Bronopoly</h1>

            <p>
                Buy properties, upgrade them, and collect rent.
                The last player with money wins.
            </p>

            <button onClick={() => setScreen("players")}>
                Create
            </button>

            <button onClick={() => setScreen("join")}>
                Join
            </button>

            <button onClick={() => setScreen("statistic")}>
                Statistic
            </button>

            <button onClick={() => setScreen("settings")}>
                Settings
            </button>

            <button onClick={() => RollDice()}/>
        </>
    );
}