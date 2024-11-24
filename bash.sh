#!/bin/bash

# Base URL
BASE_URL="http://localhost:3000"

# Student Endpoints
echo "Testing Student Endpoints..."

echo "1. GET /student/resources"
curl -X GET "$BASE_URL/student/resources" -H "Content-Type: application/json"
echo -e "\n"

echo "2. POST /student/loans"
curl -X POST "$BASE_URL/student/loans" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "student_id=1"
echo -e "\n"

echo "3. PATCH /student/update-password"
curl -X PATCH "$BASE_URL/student/update-password" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "old_password=studentpasss&new_password=newpass&student_id=1"
echo -e "\n"

# Library Agent Endpoints
echo "Testing Library Agent Endpoints..."

echo "4. GET /library-agent/overdue-loans"
curl -X GET "$BASE_URL/library-agent/overdue-loans" -H "Content-Type: application/json"
echo -e "\n"

echo "5. POST /library-agent/return-resource"
curl -X POST "$BASE_URL/library-agent/return-resource" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "loan_id=1"
echo -e "\n"

echo "6. GET /library-agent/student-profile/:student_id"
curl -X GET "$BASE_URL/library-agent/student-profile/3" -H "Content-Type: application/json"
echo -e "\n"

echo "7. POST /library-agent/assign-resource"
curl -X POST "$BASE_URL/library-agent/assign-resource" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "student_id=1&book_code=978-3-16-148410-0&loan_duration_days=15"
echo -e "\n"

# Admin Endpoints
echo "Testing Admin Endpoints..."

echo "8. POST /admin/create-student"
curl -X POST "$BASE_URL/admin/create-student" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "first_name=John&last_name=Doe&email=john.doe2@example.com&phone=1234567891&postal_address=123 Elm Street"
echo -e "\n"

echo "9. PATCH /admin/activate-card"
curl -X PATCH "$BASE_URL/admin/activate-card" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "student_id=1"
echo -e "\n"

echo "10. POST /admin/add-resource"
curl -X POST "$BASE_URL/admin/add-resource" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    --data-urlencode "book_code=978-3-16-148410-0" \
    --data-urlencode "barcode=BC005" \
    --data-urlencode "rack=1" \
    --data-urlencode "price=29.99" \
    --data-urlencode "purchase_date=2023-05-01"
echo -e "\n"

echo "All endpoint tests completed."
