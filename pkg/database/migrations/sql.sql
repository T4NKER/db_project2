CREATE TABLE IF NOT EXISTS Student (
    student_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(15) UNIQUE,
    postal_address VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS LibraryCard (
    card_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    activation_date DATE NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    resource VARCHAR(255) NOT NULL,
    FOREIGN KEY (student_id) REFERENCES Student(student_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Book (
    book_code VARCHAR(17) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    pages INT,
    author VARCHAR(255),
    publisher VARCHAR(255),
    publication_year INT,
    language VARCHAR(50),
    subject VARCHAR(100)
);
CREATE TABLE IF NOT EXISTS Book_copy (
    copy_id SERIAL PRIMARY KEY,
    book_code VARCHAR(17) NOT NULL,
    barcode VARCHAR(50) UNIQUE NOT NULL,
    rack_number INT,
    price DECIMAL(10, 2),
    purchase_date DATE,
    is_available BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (book_code) REFERENCES Book(book_code) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Loan (
    loan_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    copy_id INT NOT NULL,
    loan_date DATE NOT NULL,
    due_date DATE NOT NULL,
    return_date DATE,
    FOREIGN KEY (student_id) REFERENCES Student(student_id) ON DELETE CASCADE,
    FOREIGN KEY (copy_id) REFERENCES Book_copy(copy_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS "User" (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    user_role VARCHAR(20) NOT NULL CHECK (
        user_role IN ('Admin', 'LibraryAgent', 'Student')
    ),
    student_id INT,
    FOREIGN KEY (student_id) REFERENCES Student(student_id) ON DELETE CASCADE
);


CREATE OR REPLACE FUNCTION enforce_loan_limit() RETURNS TRIGGER AS $$ BEGIN IF NEW.student_id IS NOT NULL THEN IF (
        SELECT COUNT(*)
        FROM Loan
        WHERE student_id = NEW.student_id
    ) >= (
        CASE
            WHEN (
                SELECT COUNT(*)
                FROM LibraryCard
                WHERE student_id = NEW.student_id
            ) = 0 THEN 1 
            ELSE 5 
        END
    ) THEN RAISE EXCEPTION 'Loan limit exceeded for student ID %.',
    NEW.student_id;
END IF;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'loan_limit_trigger'
) THEN CREATE TRIGGER loan_limit_trigger BEFORE
INSERT ON Loan FOR EACH ROW EXECUTE FUNCTION enforce_loan_limit();
END IF;
END $$;
-- View for overdue books
CREATE OR REPLACE VIEW overdue_books AS
SELECT b.title,
    bc.barcode,
    l.due_date,
    s.first_name,
    s.last_name
FROM Loan l
    JOIN Book_copy bc ON l.copy_id = bc.copy_id
    JOIN Book b ON bc.book_code = b.book_code
    JOIN Student s ON l.student_id = s.student_id
WHERE l.due_date < CURRENT_DATE;
-- View for available book copies
CREATE OR REPLACE VIEW available_copies AS
SELECT b.title,
    COUNT(*) AS available_copies
FROM Book b
    JOIN Book_copy bc ON b.book_code = bc.book_code
WHERE bc.is_available = TRUE
GROUP BY b.title;
-- Procedure to activate a library card
CREATE OR REPLACE PROCEDURE activate_card(card_id INT) AS $$ BEGIN
UPDATE LibraryCard
SET status = TRUE
WHERE card_id = card_id;
END;
$$ LANGUAGE plpgsql;
-- Procedure to deactivate a library card
CREATE OR REPLACE PROCEDURE deactivate_card(card_id INT) AS $$ BEGIN
UPDATE LibraryCard
SET status = FALSE
WHERE card_id = card_id;
END;
$$ LANGUAGE plpgsql;
-- Insert Sample Data
INSERT INTO Student (
        first_name,
        last_name,
        email,
        phone,
        postal_address
    )
VALUES (
        'John',
        'Doe',
        'john.doe@example.com',
        '1234567890',
        '123 Elm Street'
    ),
    (
        'Jane',
        'Smith',
        'jane.smith@example.com',
        '0987654321',
        '456 Oak Avenue'
    ),
    (
        'Emily',
        'Johnson',
        'emily.johnson@example.com',
        '1122334455',
        '789 Pine Road'
    ) ON CONFLICT DO NOTHING;
INSERT INTO LibraryCard (
        student_id,
        activation_date,
        status,
        resource
    )
VALUES (1, '2024-01-01', TRUE, 'Computer'),
    (2, '2024-01-02', TRUE, 'Book'),
    (3, '2024-01-03', FALSE, 'Meeting room') ON CONFLICT DO NOTHING;
INSERT INTO Book (
        book_code,
        title,
        pages,
        publication_year,
        language,
        subject
    )
VALUES (
        '978-3-16-148410-0',
        'Introduction to Databases',
        300,
        2020,
        'English',
        'Computer Science'
    ),
    (
        '978-0-262-13472-9',
        'Artificial Intelligence: A Modern Approach',
        1000,
        2018,
        'English',
        'AI'
    ),
    (
        '978-1-56619-909-4',
        'Learning SQL',
        350,
        2021,
        'English',
        'Databases'
    ) ON CONFLICT DO NOTHING;
INSERT INTO Book_copy (
        book_code,
        barcode,
        rack_number,
        price,
        purchase_date,
        is_available
    )
VALUES (
        '978-3-16-148410-0',
        'BC001',
        1,
        29.99,
        '2023-05-01',
        FALSE
    ),
    (
        '978-3-16-148410-0',
        'BC002',
        1,
        29.99,
        '2023-05-01',
        TRUE
    ),
    (
        '978-0-262-13472-9',
        'BC003',
        2,
        49.99,
        '2023-06-15',
        FALSE
    ),
    (
        '978-1-56619-909-4',
        'BC004',
        3,
        39.99,
        '2023-07-20',
        TRUE
    ) ON CONFLICT DO NOTHING;
INSERT INTO Loan (student_id, copy_id, loan_date, due_date)
VALUES (1, 1, '2024-01-10', '2024-01-25'),
    (2, 3, '2024-01-12', '2024-01-27') ON CONFLICT DO NOTHING;
INSERT INTO "User" (username, password, user_role, student_id)
VALUES ('admin1', 'adminpass', 'Admin', NULL),
    (
        'libagent1',
        'libagentpass',
        'LibraryAgent',
        NULL
    ),
    ('johndoe', 'studentpass', 'Student', 1),
    ('janesmith', 'studentpass', 'Student', 2) ON CONFLICT DO NOTHING;