export default function StartScreen({ setScreen }) {
    return (
        <>
            <h1>Bronopoly</h1>

            <p>
                Buy properties, upgrade them, and collect rent.
                The last player with money wins.
            </p>

            <p>How do you want to play?</p>

            <button onClick={() => setScreen("players")}>
                Local game
            </button>

            <button onClick={() => setScreen("online")}>
                Online (coming soon)
            </button>
        </>
    );
}