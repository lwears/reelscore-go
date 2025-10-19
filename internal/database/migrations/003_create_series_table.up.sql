-- Create Serie table
CREATE TABLE "Serie" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "tmdbId" integer NOT NULL,
  "createdAt" timestamp DEFAULT now() NOT NULL,
  "updatedAt" timestamp DEFAULT now() NOT NULL,
  "title" varchar(255) NOT NULL,
  "posterPath" varchar(500),
  "firstAired" timestamp,
  "tmdbScore" numeric(3, 1) DEFAULT '0' NOT NULL,
  "score" numeric(3, 1) DEFAULT '0' NOT NULL,
  "watched" boolean DEFAULT false NOT NULL,
  "userId" uuid NOT NULL,
  CONSTRAINT "Serie_tmdbId_userId_unique" UNIQUE("tmdbId", "userId"),
  CONSTRAINT "Serie_userId_User_id_fk" FOREIGN KEY ("userId")
    REFERENCES "User"("id") ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX "idx_serie_user_id" ON "Serie"("userId");
CREATE INDEX "idx_serie_tmdb_id" ON "Serie"("tmdbId");
CREATE INDEX "idx_serie_watched" ON "Serie"("watched");
CREATE INDEX "idx_serie_title" ON "Serie"("title");
