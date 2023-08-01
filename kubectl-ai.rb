# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class KubectlAi < Formula
  desc "kubectl-ai is a kubectl plugin to generate and apply Kubernetes manifests using OpenAI GPT."
  homepage ""
  version "0.0.11"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.11/kubectl-ai_darwin_arm64.tar.gz"
      sha256 "446b8f163bb61682bb6f44770a742637209a884c094a6a4f61294f6cd8c226d1"

      def install
        bin.install "kubectl-ai"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.11/kubectl-ai_darwin_amd64.tar.gz"
      sha256 "2ee5ae161b2705aa6e382a08fa614238ad1b33f2b71be8482aac377cdecc7c8b"

      def install
        bin.install "kubectl-ai"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.11/kubectl-ai_linux_arm64.tar.gz"
      sha256 "158eca02fab65556808724e99e4f5922bd0c34105c100950458fb9a7faf1107c"

      def install
        bin.install "kubectl-ai"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.11/kubectl-ai_linux_amd64.tar.gz"
      sha256 "b991b085a59a47f6ccd0f4a907c8c082c449acadf2553938a7edf36522f6daf4"

      def install
        bin.install "kubectl-ai"
      end
    end
  end

  def caveats
    <<~EOS
      This plugin requires an OpenAI key.
    EOS
  end
end
