package connection

import "fmt"

func (r *Repository) InitSchema() error {
	_, err := r.db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return fmt.Errorf("failed to create extension: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS employee (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		username VARCHAR(50) UNIQUE NOT NULL,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create employee table: %w", err)
	}

	_, err = r.db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'organization_type') THEN
				CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC');
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to create enum type: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS organization (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		type organization_type,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create organization table: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS organization_responsible (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create organization_responsible table: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS tender (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		service_type VARCHAR(100),
		status VARCHAR(10) NOT NULL,
		organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
		creator_username VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create tender table: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS tender_version (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		description TEXT
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create tender_version table: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS bid (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		status VARCHAR(10) NOT NULL,
	  tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
		creator_username VARCHAR(50) NOT NULL,
		author_type VARCHAR(50) NOT NULL,
		author_id UUID NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create bid table: %w", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS bid_version (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		bid_id UUID REFERENCES tender(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		description TEXT
	);
`)
	if err != nil {
		return fmt.Errorf("failed to create bid_version table: %w", err)
	}

	return nil
}
