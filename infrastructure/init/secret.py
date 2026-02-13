#!/usr/bin/env python3

import os
import sys
import secrets
import argparse
from pathlib import Path
from getpass import getpass


def create_env_directory():
    """
        Creates the `.env` directory in the current working directory

        Returns:
            env_dir [Path]: the env directory represented as a path object
    """
    env_dir = Path(".env")
    env_dir.mkdir(exist_ok=True)
    return env_dir


def generate_random_secret(length: int = 32) -> str:
    """
        Generates a random URL safe secret token of a given length

        Parameters:
            length [int]: the length of token to generate (defaults to 32)

        Returns:
            token [str]: the generated token
    """
    return secrets.token_urlsafe(length)


def read_secret_as_value() -> str:
    """
        Retrieves the secret token from an input stream without exposing it

        Returns:
            secret [str]: the input secret
    """
    return getpass('Secret Value: ', echo_char='*')


def write_secret_to_file(env_dir: Path, key: str, value: str) -> None:
    """
        Writes the secret value to  `.env/{key}.txt`

        Parameters:
            env_dir [Path]: abstract path pointing to env directory
            key [str]: name of the secret to store (will be written to key.txt)
            value [str]: the secret to store
    """
    env_file = env_dir / f"{key}.txt"
    with open(env_file, "a") as f:
        f.write(f"{value}")


def main():
    """
        Entry point for the infrastructure secret initalizer

        Arguments:
            key [str]: the secret name to initialize (required)

        Flags:
            --value: source secret from interactive input
            --generate: source secret from random entropy
            --length: length of generated secret (defaults to 32)

        Outputs:
            file: a env secret file storing the specified or generated secret

        Errors:
            key name not specified
            secret source not specified
            too many sources specified
    """
    parser = argparse.ArgumentParser(
        description="Initialize secrets for the infrastructure's environment",
    )
    parser.add_argument("key", nargs="?", help="Secret key name")
    parser.add_argument(
        "--value",
        action="store_true",
        help="Read secret value interactively")
    parser.add_argument(
        "--generate",
        action="store_true",
        help="Generate random value"
    )
    parser.add_argument(
        "--length",
        type=int,
        default=32,
        help="Length for generated secrets (default: 32)"
    )

    args = parser.parse_args()

    env_dir = create_env_directory()

    if not args.key:
        print("Error: Key name is required. Use --help for usage information.")
        sys.exit(1)

    if args.value and args.generate:
        print("Error: Cannot specify both --value and --generate")
        sys.exit(1)

    if args.value:
        value = read_secret_as_value()
    elif args.generate:
        value = generate_random_secret(args.length)
        print(f"Generated random secret: {value}")
    else:
        print("Error: Must specify secret source using --value or --generate")
        sys.exit(1)

    write_secret_to_file(env_dir, args.key, value)


if __name__ == "__main__":
    main()
