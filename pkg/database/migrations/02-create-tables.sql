CREATE TABLE IF NOT EXISTS Student (
    student_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(15) UNIQUE,
    postal_address VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS Resource (
    resource_id SERIAL PRIMARY KEY,
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN ('Book', 'Computer', 'Room')),
    description VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS LibraryCard (
    card_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    activation_date DATE NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    resource_id INT NOT NULL,
    FOREIGN KEY (student_id) REFERENCES Student(student_id) ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES Resource(resource_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Book (
    book_code VARCHAR(17) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    pages INT
);
CREATE TABLE IF NOT EXISTS Book_Language (
    book_code VARCHAR(17) NOT NULL,
    language VARCHAR(50) NOT NULL,
    FOREIGN KEY (book_code) REFERENCES Book(book_code) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Author (
    author_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    birth_date DATE,
    nationality VARCHAR(50),
    biography TEXT
);
CREATE TABLE IF NOT EXISTS Book_Author (
    book_code VARCHAR(17) NOT NULL,
    author_id INT NOT NULL,
    FOREIGN KEY (book_code) REFERENCES Book(book_code) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES Author(author_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Publisher (
    publisher_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS "Subject" (
    subject_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS Book_Publisher (
    book_code VARCHAR(17) NOT NULL,
    publisher_id INT NOT NULL,
    FOREIGN KEY (book_code) REFERENCES Book(book_code) ON DELETE CASCADE,
    FOREIGN KEY (publisher_id) REFERENCES Publisher(publisher_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Book_Subject (
    book_code VARCHAR(17) NOT NULL,
    subject_id INT NOT NULL,
    FOREIGN KEY (book_code) REFERENCES Book(book_code) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES "Subject"(subject_id) ON DELETE CASCADE
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
    return_date DATE DEFAULT NULL,
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
CREATE OR REPLACE FUNCTION enforce_loan_limit() RETURNS TRIGGER AS $$
DECLARE active_loans_count INT;
active_loans_details TEXT;
BEGIN -- Count the number of active loans for the student
SELECT COUNT(*) INTO active_loans_count
FROM Loan
WHERE student_id = NEW.student_id
    AND return_date IS NULL;
-- Retrieve details of active loans (for debugging purposes)
SELECT STRING_AGG(loan_id::TEXT || '-' || copy_id::TEXT, ', ') INTO active_loans_details
FROM Loan
WHERE student_id = NEW.student_id
    AND return_date IS NULL;
-- Check if the student is not registered
IF (
    SELECT COUNT(*)
    FROM LibraryCard
    WHERE student_id = NEW.student_id
) = 0 THEN -- Raise exception if the limit is exceeded
IF active_loans_count >= 1 THEN RAISE EXCEPTION 'Non-registered student % has % active loan(s): [%]. Only 1 book allowed.',
NEW.student_id,
active_loans_count,
active_loans_details;
END IF;
ELSE -- Raise exception if the limit is exceeded for registered students
IF active_loans_count >= 5 THEN RAISE EXCEPTION 'Registered student % has % active loan(s): [%]. Up to 5 books allowed.',
NEW.student_id,
active_loans_count,
active_loans_details;
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
CREATE OR REPLACE VIEW available_copies AS
SELECT b.title,
    STRING_AGG(
        DISTINCT a.first_name || ' ' || a.last_name,
        ', '
    ) AS authors,
    STRING_AGG(DISTINCT bl.language, ', ') AS languages,
    p.name AS publisher,
    COUNT(*) AS available_copies
FROM Book b
    JOIN Book_copy bc ON b.book_code = bc.book_code
    LEFT JOIN Book_Author ba ON b.book_code = ba.book_code
    LEFT JOIN Author a ON ba.author_id = a.author_id
    LEFT JOIN Book_Language bl ON b.book_code = bl.book_code
    LEFT JOIN Book_Publisher bp ON b.book_code = bp.book_code
    LEFT JOIN Publisher p ON bp.publisher_id = p.publisher_id
WHERE bc.is_available = TRUE
GROUP BY b.title,
    p.name;
CREATE OR REPLACE PROCEDURE toggle_card_status(student_id INT) AS $$ BEGIN -- 
    IF NOT EXISTS (
        SELECT 1
        FROM LibraryCard
        WHERE LibraryCard.student_id = toggle_card_status.student_id
    ) THEN RAISE EXCEPTION 'No library card found for student_id %',
    student_id;
END IF;
UPDATE LibraryCard
SET status = NOT status
WHERE LibraryCard.student_id = toggle_card_status.student_id;
END;
$$ LANGUAGE plpgsql;
INSERT INTO Resource (resource_type, description)
VALUES ('Book', 'General books available for loan'),
    ('Computer', 'Library computers for student use'),
    ('Room', 'Study rooms available for reservation');
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
INSERT INTO LibraryCard (student_id, activation_date, status, resource_id)
VALUES (1, '2024-01-01', TRUE, 1),
    (2, '2024-01-02', TRUE, 2);
INSERT INTO Book (book_code, title, pages)
VALUES (
        '978-3-16-148410-0',
        'Introduction to Databases',
        300
    ),
    (
        '978-0-262-13472-9',
        'Artificial Intelligence: A Modern Approach',
        1000
    ),
    ('978-1-56619-909-4', 'Learning SQL', 350);
INSERT INTO Book_Language (book_code, language)
VALUES ('978-3-16-148410-0', 'English'),
    ('978-3-16-148410-0', 'Spanish'),
    ('978-0-262-13472-9', 'English'),
    ('978-1-56619-909-4', 'English'),
    ('978-1-56619-909-4', 'French');
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
INSERT INTO Author (
        first_name,
        last_name,
        birth_date,
        nationality,
        biography
    )
VALUES (
        'John',
        'Stone',
        '1975-05-15',
        'American',
        'Expert in Databases and Data Structures'
    ),
    (
        'Stuart',
        'Russell',
        '1962-06-22',
        'British',
        'Renowned for AI research and books'
    ),
    (
        'Mary',
        'Robertson',
        '1980-11-03',
        'Canadian',
        'Author specializing in SQL and relational databases'
    );
INSERT INTO Publisher (name)
VALUES ('Books n Pieces'),
    ('TechPress'),
    ('Global Publishing');
INSERT INTO "Subject" (name)
VALUES ('Databases'),
    ('Artificial Intelligence'),
    ('SQL');
INSERT INTO Book_Author (book_code, author_id)
VALUES ('978-3-16-148410-0', 1),
    ('978-0-262-13472-9', 2),
    ('978-1-56619-909-4', 3);
INSERT INTO Book_Publisher (book_code, publisher_id)
VALUES ('978-3-16-148410-0', 1),
    ('978-0-262-13472-9', 2),
    ('978-1-56619-909-4', 3);
INSERT INTO Book_Subject (book_code, subject_id)
VALUES ('978-3-16-148410-0', 1),
    ('978-0-262-13472-9', 2),
    ('978-1-56619-909-4', 3);