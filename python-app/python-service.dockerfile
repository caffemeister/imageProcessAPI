# Use the official Python image from the Docker registry
FROM python:3.9-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the Poetry configuration and install dependencies
COPY pyproject.toml poetry.lock /app/

# Install Poetry and dependencies
RUN pip install poetry && poetry install --no-dev

# Copy the application code into the container
COPY . /app/

# Expose the port that the Flask app will run on
EXPOSE 8000

# Command to run the Flask app inside the container
CMD ["poetry", "run", "flask", "run", "--host", "0.0.0.0", "--port", "8000"]
