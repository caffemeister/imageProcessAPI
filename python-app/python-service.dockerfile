# Use the official Python image from the Docker registry
FROM python:3.12

# Set the working directory inside the container
WORKDIR /app

# Install Poetry globally
RUN pip install poetry

# Configure Poetry to install dependencies inside the project (not globally)
ENV POETRY_VIRTUALENVS_IN_PROJECT=true
ENV POETRY_NO_INTERACTION=1

# Copy Poetry configuration files first (for better caching)
COPY pyproject.toml poetry.lock /app/

# Install dependencies inside the virtual environment
RUN poetry install --no-root

# Copy the rest of the application code into the container
COPY . /app/

# Expose the port that the Flask app will run on
EXPOSE 8000

# Ensure the virtual environment is activated before running the app
CMD ["poetry", "run", "python", "main.py"]
