CREATE TABLE player (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    points NUMERIC NOT NULL, 
    wins NUMERIC NOT NULL,
    words_guessed NUMERIC NOT NULL,
    drawings_guessed NUMERIC NOT NULL
    avatar TEXT NOT NULL
);