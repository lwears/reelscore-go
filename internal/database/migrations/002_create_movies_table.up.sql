-- Create Movie table
CREATE TABLE "Movie" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "tmdbId" integer NOT NULL,
  "createdAt" timestamp DEFAULT now() NOT NULL,
  "updatedAt" timestamp DEFAULT now() NOT NULL,
  "title" varchar(255) NOT NULL,
  "posterPath" varchar(500),
  "releaseDate" timestamp,
  "tmdbScore" numeric(3, 1) DEFAULT '0' NOT NULL,
  "score" numeric(3, 1) DEFAULT '0' NOT NULL,
  "watched" boolean DEFAULT false NOT NULL,
  "userId" uuid NOT NULL,
  CONSTRAINT "Movie_tmdbId_userId_unique" UNIQUE("tmdbId", "userId"),
  CONSTRAINT "Movie_userId_User_id_fk" FOREIGN KEY ("userId")
    REFERENCES "User"("id") ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX "idx_movie_user_id" ON "Movie"("userId");
CREATE INDEX "idx_movie_tmdb_id" ON "Movie"("tmdbId");
CREATE INDEX "idx_movie_watched" ON "Movie"("watched");
CREATE INDEX "idx_movie_title" ON "Movie"("title");
