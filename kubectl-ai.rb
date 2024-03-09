# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class KubectlAi < Formula
  desc "kubectl-ai is a kubectl plugin to generate and apply Kubernetes manifests using OpenAI GPT."
  homepage ""
  version "0.0.12"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.12/kubectl-ai_darwin_arm64.tar.gz"
      sha256 "4adc6afb513bb3849e30d2900d5eda513c144973eec48a4ac6ee3e5d8293769f"

      def install
        bin.install "kubectl-ai"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.12/kubectl-ai_darwin_amd64.tar.gz"
      sha256 "d2efe829420487d4772adda8319ae74111cf2d6b40cd52ffe873ec4011544d32"

      def install
        bin.install "kubectl-ai"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.12/kubectl-ai_linux_amd64.tar.gz"
      sha256 "6f390023a8fe46a5237a15e934be3189230896dff64bc2a49cf4c8f1df59d690"

      def install
        bin.install "kubectl-ai"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/sozercan/kubectl-ai/releases/download/v0.0.12/kubectl-ai_linux_arm64.tar.gz"
      sha256 "2a8897c4b0fca40a1e980abf489713de024ae80a93c247afd39697c5b42edf59"

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
