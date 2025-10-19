-- Create Provider enum type
CREATE TYPE "Provider" AS ENUM ('GITHUB', 'GOOGLE');

-- Create User table
CREATE TABLE "User" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "providerId" varchar(255) NOT NULL UNIQUE,
  "provider" "Provider" NOT NULL,
  "email" varchar(255) NOT NULL,
  "name" varchar(255) NOT NULL,
  "createdAt" timestamp DEFAULT now() NOT NULL,
  "updatedAt" timestamp DEFAULT now() NOT NULL
);

-- Create index on providerId for faster lookups
CREATE INDEX "idx_user_provider_id" ON "User"("providerId");

-- Create index on email for faster lookups
CREATE INDEX "idx_user_email" ON "User"("email");
