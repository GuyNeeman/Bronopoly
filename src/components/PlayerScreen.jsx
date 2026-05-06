import { Slider } from "@mui/material";

export default function PlayerScreen({ playerCount, setPlayerCount, next, back }) {
    return (
        <>
            <h2>Players</h2>

            <p>How many people are playing?</p>

            <Slider
                value={playerCount}
                min={2}
                max={5}
                valueLabelDisplay="auto"
                onChange={(e, val) => setPlayerCount(val)}
            />

            <p><strong>{playerCount}</strong> players</p>

            <button onClick={back}>Back</button>
            <button onClick={next}>Next</button>
        </>
    );
}