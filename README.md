# Zarf Enhancement Proposals (ZEPs)
Formal documents proposing significant Zarf changes, features, or enhancements.  ZEPs are premised off Kuberetes Enhancement Proposals (KEPs).

## Purpose
Provide a structured way for Zarf community contributors to suggest, discuss, and track new features or changes to the Zarf project. ZEPs contribute to Zarf transparency, community input, and alignment with [Roadmap](https://docs.zarf.dev/roadmap) goals and priorities.

## Desired Outcome
Improved visibility for proposed changes, better coordination across the Zarf community, and a clear vision of what Zarf does and does not support as it evolves.⁠

## How ZEPs Work
- Contributors draft ZEPs as GitHub pull requests in the zarf-dev/proposals repository.

- ZEPs are discussed and reviewed by the Zarf maintainers and the community, via the Kubernetes Slack chat, public [Zarf Community Meetings](https://www.google.com/url?q=https://zoom-lfx.platform.linuxfoundation.org/meeting/97461829237?password%3Dadd48ad5-fc07-4951-96d2-531b72d2a5dc&sa=D&source=calendar&ust=1747595030129732&usg=AOvVaw3mFLYGKyTC_8Q97lGnHegX), or in the Github zarf-dev/proposals ZEP Pull Request itself. Key topics include clarifying the problem, outlining the proposed solution, and identifying related dependencies or design questions.⁠
⁠​
- Once reviewed, a ZEP can be accepted, rejected, or returned for revisions. Accepted ZEPs move forward into development, with tickets or issues tracked in the [Zarf GitHub Project Board](https://github.com/orgs/zarf-dev/projects/1).⁠
⁠
### ZEP Phases
- provisional: The ZEP has been proposed and is actively being defined. This is the starting state while the ZEP is being fleshed out and actively defined and discussed.
- implementable: The approvers have approved this ZEP for implementation.
- implemented: The ZEP has been implemented and is no longer actively changed.
- deferred: The ZEP is proposed but not actively being worked on.
- rejected: The approvers and authors have decided that this ZEP is not moving forward. The ZEP is kept around as a historical document.
- withdrawn: The authors have withdrawn the ZEP.
- replaced: The ZEP has been replaced by a new ZEP. The superseded-by metadata value should point to the new ZEP.

### ZEP Status:
- A ZEP's status (provisional, deferred, implemented, etc.) can be found in the ZEP's **zep.yaml** under "status."

## Getting started
Follow the guide at the top of the [ZEP Template](NNNN-zep-template/zep.yaml) to get started on your ZEP
